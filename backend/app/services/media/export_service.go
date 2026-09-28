package media

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/support/path"
	"goravel/app/facades"
	"goravel/app/models"
	auditservices "goravel/app/services/audit"
	storageservices "goravel/app/services/storage"
)

const (
	ExportStatusQueued  = "queued"
	ExportStatusRunning = "running"
	ExportStatusReady   = "ready"
	ExportStatusFailed  = "failed"
)

type MediaExportService struct{}

func NewMediaExportService() *MediaExportService { return &MediaExportService{} }

func (s *MediaExportService) Create(ctx context.Context, userID uint) (models.MediaExportJob, error) {
	if userID == 0 {
		return models.MediaExportJob{}, errors.New("invalid user")
	}
	var active models.MediaExportJob
	if err := facades.Orm().Query().Where("user_id = ? AND status IN (?, ?)", userID, ExportStatusQueued, ExportStatusRunning).First(&active); err == nil && active.ID > 0 {
		return active, nil
	}
	job := models.MediaExportJob{UserID: userID, Status: ExportStatusQueued}
	if err := facades.Orm().Query().Create(&job); err != nil {
		return models.MediaExportJob{}, err
	}
	go func() {
		workerCtx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()
		_ = s.run(workerCtx, job.ID)
	}()
	_ = auditservices.NewAuditService().Record(userID, "media.export.create", map[string]any{"export_id": job.ID})
	return job, nil
}

func (s *MediaExportService) ListPage(userID uint, page, perPage int) ([]models.MediaExportJob, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	var jobs []models.MediaExportJob
	var total int64
	err := facades.Orm().Query().Where("user_id = ?", userID).OrderByDesc("id").Paginate(page, perPage, &jobs, &total)
	return jobs, total, err
}

func (s *MediaExportService) Find(userID, id uint) (models.MediaExportJob, error) {
	var job models.MediaExportJob
	if err := facades.Orm().Query().Where("user_id = ? AND id = ?", userID, id).First(&job); err != nil {
		return models.MediaExportJob{}, err
	}
	return job, nil
}

func (s *MediaExportService) DownloadPath(job models.MediaExportJob) (string, error) {
	if job.Status != ExportStatusReady || job.StoragePath == "" {
		return "", errors.New("export is not ready")
	}
	clean := filepath.Clean(job.StoragePath)
	root := filepath.Clean(exportRoot())
	if filepath.Dir(clean) != root {
		return "", errors.New("export path is invalid")
	}
	if _, err := os.Stat(clean); err != nil {
		return "", err
	}
	return clean, nil
}

func exportRoot() string {
	if value := strings.TrimSpace(os.Getenv("FASTIMG_EXPORT_STORAGE_PATH")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Clean(path.Storage("exports"))
}

func (s *MediaExportService) run(ctx context.Context, id uint) error {
	var job models.MediaExportJob
	if err := facades.Orm().Query().Find(&job, id); err != nil {
		return err
	}
	now := time.Now().UTC()
	_, _ = facades.Orm().Query().Table("media_export_jobs").Where("id = ?", id).Update(map[string]any{"status": ExportStatusRunning, "updated_at": now})
	if err := os.MkdirAll(exportRoot(), 0o700); err != nil {
		return s.fail(id, "EXPORT_STORAGE_UNAVAILABLE", err)
	}
	pathName := filepath.Join(exportRoot(), fmt.Sprintf("fastimg-media-export-%d.zip", id))
	file, err := os.Create(pathName)
	if err != nil {
		return s.fail(id, "EXPORT_STORAGE_UNAVAILABLE", err)
	}
	archive := zip.NewWriter(file)
	count, err := s.writeMedia(ctx, archive, job.UserID)
	if err == nil {
		meta := map[string]any{"user_id": job.UserID, "generated_at": now, "format": "fastimg-media-export-v1"}
		if err = writeZipJSON(archive, "manifest.json", meta); err == nil {
			err = archive.Close()
		}
	} else {
		_ = archive.Close()
	}
	_ = file.Close()
	if err != nil {
		_ = os.Remove(pathName)
		return s.fail(id, "EXPORT_FAILED", err)
	}
	info, err := os.Stat(pathName)
	if err != nil {
		return s.fail(id, "EXPORT_FAILED", err)
	}
	completed := time.Now().UTC()
	_, err = facades.Orm().Query().Table("media_export_jobs").Where("id = ?", id).Update(map[string]any{"status": ExportStatusReady, "file_name": filepath.Base(pathName), "storage_path": pathName, "size_bytes": info.Size(), "media_count": count, "updated_at": completed})
	if err == nil {
		_ = auditservices.NewAuditService().Record(job.UserID, "media.export.ready", map[string]any{"export_id": id, "media_count": count})
	}
	return err
}

func (s *MediaExportService) writeMedia(ctx context.Context, archive *zip.Writer, userID uint) (int64, error) {
	var assets []models.MediaAsset
	if err := facades.Orm().Query().Where("user_id = ? AND status = ?", userID, "ready").OrderBy("id", "asc").Get(&assets); err != nil {
		return 0, err
	}
	disk := facades.Storage().Disk("fastimg")
	registry := storageservices.NewRuntimeRegistry(disk)
	var links strings.Builder
	var count int64
	for _, asset := range assets {
		if err := ctx.Err(); err != nil {
			return count, err
		}
		var variant models.MediaVariant
		if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", asset.ID, "original", "ready").First(&variant); err != nil {
			continue
		}
		var object models.StorageObject
		if err := facades.Orm().Query().Find(&object, variant.StorageObjectID); err != nil {
			return count, err
		}
		var connection models.StorageConnection
		if err := facades.Orm().Query().Find(&connection, object.StorageConnectionID); err != nil {
			return count, err
		}
		provider, err := registry.ProviderFor(connection)
		if err != nil {
			return count, err
		}
		body, err := provider.Get(ctx, object.ObjectKey)
		if err != nil {
			return count, err
		}
		name := fmt.Sprintf("media/%d-%s", asset.ID, safeExportName(asset.OriginalName))
		if err := writeZipBytes(archive, name, body); err != nil {
			return count, err
		}
		metadata := map[string]any{"id": asset.ID, "name": asset.OriginalName, "content_type": asset.ContentType, "size_bytes": asset.SizeBytes, "width": asset.Width, "height": asset.Height, "visibility": asset.Visibility, "expires_at": asset.ExpiresAt, "created_at": asset.CreatedAt}
		if err := writeZipJSON(archive, fmt.Sprintf("metadata/%d.json", asset.ID), metadata); err != nil {
			return count, err
		}
		links.WriteString(fmt.Sprintf("%d\t%s\t/api/v1/media/%d/content?variant=original\n", asset.ID, asset.OriginalName, asset.ID))
		count++
	}
	if err := writeZipBytes(archive, "links.tsv", []byte(links.String())); err != nil {
		return count, err
	}
	if err := s.writeUserData(archive, userID); err != nil {
		return count, err
	}
	return count, nil
}

// writeUserData deliberately exports metadata only for credentials. Token
// secrets and hashes never leave the database, while collection relations are
// included so an archive can be used when moving to another FastImg instance.
func (s *MediaExportService) writeUserData(archive *zip.Writer, userID uint) error {
	var folders []map[string]any
	if err := facades.Orm().Query().Table("folders").Where("user_id = ?", userID).OrderBy("id", "asc").Get(&folders); err != nil {
		return err
	}
	if err := writeZipJSON(archive, "folders.json", folders); err != nil {
		return err
	}
	var albums []map[string]any
	if err := facades.Orm().Query().Table("albums").Where("user_id = ?", userID).OrderBy("id", "asc").Get(&albums); err != nil {
		return err
	}
	if err := writeZipJSON(archive, "albums.json", albums); err != nil {
		return err
	}
	var albumMedia []map[string]any
	if err := facades.Orm().Query().Table("album_media").Where("user_id = ?", userID).OrderBy("id", "asc").Get(&albumMedia); err != nil {
		return err
	}
	if err := writeZipJSON(archive, "album-media.json", albumMedia); err != nil {
		return err
	}
	var tokens []map[string]any
	if err := facades.Orm().Query().Table("api_tokens").Where("user_id = ?", userID).OrderBy("id", "asc").Get(&tokens); err != nil {
		return err
	}
	for _, token := range tokens {
		for _, key := range []string{"token", "secret", "token_hash", "hash"} {
			delete(token, key)
		}
	}
	return writeZipJSON(archive, "token-metadata.json", tokens)
}

func (s *MediaExportService) fail(id uint, code string, err error) error {
	message := "export failed"
	if err != nil {
		message = err.Error()
	}
	_, _ = facades.Orm().Query().Table("media_export_jobs").Where("id = ?", id).Update(map[string]any{"status": ExportStatusFailed, "error_code": code, "error_message": message, "updated_at": time.Now().UTC()})
	return err
}

func writeZipBytes(archive *zip.Writer, name string, body []byte) error {
	writer, err := archive.Create(name)
	if err != nil {
		return err
	}
	_, err = writer.Write(body)
	return err
}
func writeZipJSON(archive *zip.Writer, name string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeZipBytes(archive, name, body)
}
func safeExportName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	if name == "." || name == "" {
		return "media.bin"
	}
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r < 32 {
			return '_'
		}
		return r
	}, name)
}
