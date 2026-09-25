package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type MediaAsset struct {
	orm.Model
	UserID               uint       `json:"user_id"`
	FolderID             *uint      `json:"folder_id" gorm:"column:folder_id;->"`
	OriginalName         string     `json:"original_name"`
	ContentType          string     `json:"content_type"`
	Format               string     `json:"format"`
	SizeBytes            int64      `json:"size_bytes"`
	SHA256               string     `json:"sha256"`
	Width                int64      `json:"width"`
	Height               int64      `json:"height"`
	Status               string     `json:"status"`
	Visibility           string     `json:"visibility"`
	ModerationStatus     string     `json:"moderation_status"`
	DiscoverySubmittedAt *time.Time `json:"discovery_submitted_at"`
	DeletedAt            *time.Time `json:"deleted_at"`
}

type MediaVariant struct {
	orm.Model
	MediaAssetID    uint   `json:"media_asset_id"`
	StorageObjectID uint   `json:"storage_object_id"`
	Name            string `json:"name"`
	Width           int64  `json:"width"`
	Height          int64  `json:"height"`
	Status          string `json:"status"`
}

type StorageObject struct {
	orm.Model
	Provider    string `json:"provider"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	Status      string `json:"status"`
}

type UploadSession struct {
	orm.Model
	UserID         uint   `json:"user_id"`
	MediaAssetID   uint   `json:"media_asset_id"`
	IdempotencyKey string `json:"idempotency_key"`
	SHA256         string `json:"sha256"`
	SizeBytes      int64  `json:"size_bytes"`
	ReservedBytes  int64  `json:"reserved_bytes"`
	Status         string `json:"status"`
	ErrorCode      string `json:"error_code"`
}
