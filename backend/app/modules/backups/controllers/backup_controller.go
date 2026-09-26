package controllers

import (
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"

	adminmiddleware "goravel/app/http/middleware"
)

const (
	BackupManagePermission   = "admin.backups.manage"
	BackupDownloadPermission = "admin.backups.download"
)

type BackupController struct{}

func NewBackupController() *BackupController { return &BackupController{} }

func (c *BackupController) List(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func (c *BackupController) Create(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func (c *BackupController) Show(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func (c *BackupController) Download(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func (c *BackupController) Delete(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func (c *BackupController) Validate(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func (c *BackupController) Restore(ctx httpcontract.Context) httpcontract.Response {
	return notReady(ctx)
}

func notReady(ctx httpcontract.Context) httpcontract.Response {
	return adminmiddleware.APIError(ctx, http.StatusNotImplemented, "BACKUP_FEATURE_NOT_READY")
}

