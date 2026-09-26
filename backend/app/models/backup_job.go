package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type BackupJob struct {
	orm.Model
	Kind          string     `json:"kind"`
	Status        string     `json:"status"`
	FileName      string     `json:"file_name"`
	StoragePath   string     `json:"-"`
	ManifestJSON  string     `json:"manifest_json"`
	SizeBytes     int64      `json:"size_bytes"`
	ErrorCode     string     `json:"error_code"`
	ErrorMessage  string     `json:"error_message"`
	CreatedBy     uint       `json:"created_by"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

