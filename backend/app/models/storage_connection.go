package models

import (
	"strings"
	"time"

	"github.com/goravel/framework/database/orm"
)

const (
	StorageProviderLocal        = "local"
	StorageProviderCloudflareR2 = "cloudflare_r2"
	StorageProviderAliyunOSS    = "aliyun_oss"
	StorageProviderTencentCOS   = "tencent_cos"

	StorageConnectionDisabled   = "disabled"
	StorageConnectionIncomplete = "incomplete"
	StorageConnectionHealthy    = "healthy"
	StorageConnectionDegraded   = "degraded"
	StorageConnectionError      = "error"
)

type StorageConnection struct {
	orm.Model
	ProviderCode        string     `json:"provider_code"`
	Name                string     `json:"name"`
	Enabled             bool       `json:"enabled"`
	IsPrimary           bool       `json:"is_primary"`
	Status              string     `json:"status"`
	ConfigEncrypted     string     `json:"-"`
	PublicBaseURL       string     `json:"public_base_url"`
	PathPrefix          string     `json:"path_prefix"`
	DefaultVisibility   string     `json:"default_visibility"`
	SignedURLTTLSeconds int        `json:"signed_url_ttl_seconds"`
	LastCheckedAt       *time.Time `json:"last_checked_at"`
	LastSuccessAt       *time.Time `json:"last_success_at"`
	LastErrorCode       string     `json:"last_error_code"`
	LastErrorMessage    string     `json:"last_error_message"`
}

func IsStorageProviderCode(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case StorageProviderLocal, StorageProviderCloudflareR2, StorageProviderAliyunOSS, StorageProviderTencentCOS:
		return true
	default:
		return false
	}
}

func IsStorageConnectionStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case StorageConnectionDisabled, StorageConnectionIncomplete, StorageConnectionHealthy, StorageConnectionDegraded, StorageConnectionError:
		return true
	default:
		return false
	}
}
