package models

import "github.com/goravel/framework/database/orm"

type MediaExportJob struct {
	orm.Model
	UserID       uint   `json:"user_id"`
	Status       string `json:"status"`
	FileName     string `json:"file_name"`
	StoragePath  string `json:"-"`
	SizeBytes    int64  `json:"size_bytes"`
	MediaCount   int64  `json:"media_count"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
