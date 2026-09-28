package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
	"goravel/app/models"
	auditservices "goravel/app/services/audit"
	storageservices "goravel/app/services/storage"
)

const (
	SecurityStatusClear             = "clear"
	SecurityStatusHighRisk          = "high_risk"
	SecurityReasonMagicMismatch     = "magic_mismatch"
	SecurityReasonInvalidImage      = "invalid_image"
	SecurityReasonPixelLimit        = "pixel_limit"
	SecurityReasonSuspiciousPayload = "suspicious_payload"
)

var ErrSecurityScanUnavailable = errors.New("security scan unavailable")

type SecurityScanResult struct {
	Status   string
	Reason   string
	Score    int
	Findings []string
}

// AnalyzeSecurity is intentionally deterministic and provider-free. It is the
// local safety baseline; optional malware/NSFW providers can add findings later
// without changing the upload state machine.
func AnalyzeSecurity(filename, declaredContentType string, content []byte, maxPixels int64) SecurityScanResult {
	result := SecurityScanResult{Status: SecurityStatusClear, Score: 0, Findings: []string{}}
	if len(content) == 0 {
		return highRisk(SecurityReasonInvalidImage, "empty_content")
	}
	format, ok := formatForExtension(filename)
	if !ok || !matchesMagic(format, content) {
		return highRisk(SecurityReasonMagicMismatch, "content_signature_mismatch")
	}
	if declaredContentType != "" {
		declared, _, _ := strings.Cut(strings.ToLower(declaredContentType), ";")
		declared = strings.TrimSpace(declared)
		if expected, supported := formatForMIME(declared); !supported || expected != format {
			return highRisk(SecurityReasonMagicMismatch, "declared_content_type_mismatch")
		}
	}
	if suspiciousImagePayload(content) {
		return highRisk(SecurityReasonSuspiciousPayload, "executable_marker")
	}
	if config, _, err := image.DecodeConfig(bytes.NewReader(content)); err != nil {
		return highRisk(SecurityReasonInvalidImage, "decode_failed")
	} else if maxPixels > 0 && int64(config.Width) > maxPixels/int64(max(1, config.Height)) {
		return highRisk(SecurityReasonPixelLimit, "decoded_pixel_limit")
	}
	return result
}

func highRisk(reason, finding string) SecurityScanResult {
	return SecurityScanResult{Status: SecurityStatusHighRisk, Reason: reason, Score: 100, Findings: []string{finding}}
}

func matchesMagic(format string, content []byte) bool {
	switch format {
	case "jpeg":
		return len(content) >= 3 && content[0] == 0xff && content[1] == 0xd8 && content[2] == 0xff
	case "png":
		return len(content) >= 8 && bytes.Equal(content[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	case "gif":
		return len(content) >= 6 && (bytes.Equal(content[:6], []byte("GIF87a")) || bytes.Equal(content[:6], []byte("GIF89a")))
	default:
		return false
	}
}

func suspiciousImagePayload(content []byte) bool {
	trimmed := bytes.ToLower(content)
	for _, marker := range [][]byte{[]byte("<?php"), []byte("<script"), []byte("powershell -"), []byte("cmd.exe /c")} {
		if bytes.Contains(trimmed, marker) {
			return true
		}
	}
	return len(content) >= 2 && content[0] == 'M' && content[1] == 'Z'
}

type SecurityScanJob struct{}

const SecurityScanJobSignature = "fastimg.media.security-scan"

func (*SecurityScanJob) Signature() string { return SecurityScanJobSignature }

func (*SecurityScanJob) Handle(args ...any) error {
	if len(args) == 0 {
		return errors.New("media id is required")
	}
	mediaID, ok := args[0].(uint)
	if !ok || mediaID == 0 {
		return errors.New("invalid media id")
	}
	return NewSecurityScanService().Run(context.Background(), mediaID)
}

func (*SecurityScanJob) ShouldRetry(_ error, attempt int) (bool, time.Duration) {
	if attempt >= 3 {
		return false, 0
	}
	return true, time.Duration(attempt+1) * time.Minute
}

func DispatchSecurityScan(mediaID uint) error {
	if mediaID == 0 {
		return errors.New("invalid media id")
	}
	if !facades.Schema().HasTable("media_security_scans") {
		return ErrSecurityScanUnavailable
	}
	exists, err := facades.Orm().Query().Table("media_security_scans").Where("media_asset_id = ?", mediaID).Exists()
	if err != nil {
		return err
	}
	if !exists {
		if err := facades.Orm().Query().Table("media_security_scans").Create(&models.MediaSecurityScan{MediaAssetID: mediaID, Status: models.SecurityScanQueued, Provider: "local"}); err != nil {
			return err
		}
	}
	return facades.Queue().Job(&SecurityScanJob{}, []queue.Arg{{Value: mediaID, Type: "uint"}}).OnQueue("media-security").Dispatch()
}

type SecurityScanService struct{}

func NewSecurityScanService() *SecurityScanService { return &SecurityScanService{} }

func (s *SecurityScanService) Run(ctx context.Context, mediaID uint) error {
	if !facades.Schema().HasTable("media_security_scans") {
		return ErrSecurityScanUnavailable
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaID, "ready").First(&asset); err != nil {
		return err
	}
	var variant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", mediaID, "original", "ready").First(&variant); err != nil {
		return err
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Find(&object, variant.StorageObjectID); err != nil {
		return err
	}
	var connection models.StorageConnection
	if err := facades.Orm().Query().Find(&connection, object.StorageConnectionID); err != nil {
		return err
	}
	provider, err := storageservices.NewRuntimeRegistry(facades.Storage().Disk("fastimg")).ProviderFor(connection)
	if err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Table("media_security_scans").Where("media_asset_id = ?", mediaID).Update(map[string]any{"status": models.SecurityScanRunning, "started_at": time.Now().UTC(), "updated_at": time.Now().UTC()}); err != nil {
		return err
	}
	body, err := provider.Get(ctx, object.ObjectKey)
	if err != nil {
		return err
	}
	result := AnalyzeSecurity(asset.OriginalName, asset.ContentType, body, 20_000_000)
	findings, _ := json.Marshal(result.Findings)
	now := time.Now().UTC()
	status := result.Status
	if status == SecurityStatusClear {
		status = models.SecurityScanClear
	}
	if _, err := facades.Orm().Query().Table("media_security_scans").Where("media_asset_id = ?", mediaID).Update(map[string]any{"status": status, "reason": result.Reason, "risk_score": result.Score, "findings_json": string(findings), "completed_at": now, "updated_at": now}); err != nil {
		return err
	}
	if result.Status == SecurityStatusHighRisk {
		if _, err := facades.Orm().Query().Table("media_assets").Where("id = ? AND status = ?", mediaID, "ready").Update(map[string]any{"visibility": "private", "moderation_status": "manual_review", "updated_at": now}); err != nil {
			return err
		}
		if facades.Schema().HasTable("media_reports") {
			_ = facades.Orm().Query().Table("media_reports").Create(map[string]any{"media_asset_id": mediaID, "reason": "automated_security", "description": result.Reason, "status": "pending", "created_at": now, "updated_at": now})
		}
		_ = auditservices.NewAuditService().Record(asset.UserID, "moderation.security_scan.hide", map[string]any{"media_id": mediaID, "reason": result.Reason})
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
