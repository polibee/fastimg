package controllers

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	backupservices "goravel/app/services/backups"
)

const (
	BackupManagePermission   = "admin.backups.manage"
	BackupDownloadPermission = "admin.backups.download"
)

const maxBackupUploadBytes int64 = 10 << 30

type BackupController struct {
	service *backupservices.BackupService
	restore *backupservices.RestoreService
}

func NewBackupController() *BackupController {
	service := backupservices.NewBackupService()
	return &BackupController{service: service, restore: backupservices.NewRestoreServiceWithDependencies(service, nil)}
}

func (c *BackupController) List(ctx httpcontract.Context) httpcontract.Response {
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Request().Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	jobs, total, err := c.service.List(page, perPage)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "BACKUPS_UNAVAILABLE")
	}
	for index := range jobs {
		jobs[index].StoragePath = ""
		jobs[index].ErrorMessage = safeErrorMessage(jobs[index].ErrorMessage)
	}
	lastPage := int64(1)
	if total > 0 {
		lastPage = (total + int64(perPage) - 1) / int64(perPage)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": jobs, "meta": httpcontract.Json{"page": page, "per_page": perPage, "total": total, "last_page": lastPage}})
}

func (c *BackupController) Create(ctx httpcontract.Context) httpcontract.Response {
	operatorID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	job, err := c.service.Create(ctx.Context(), operatorID)
	if errors.Is(err, backupservices.ErrBackupActive) {
		return adminmiddleware.APIError(ctx, http.StatusConflict, "BACKUP_ALREADY_RUNNING")
	}
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "BACKUP_CREATE_FAILED")
	}
	return ctx.Response().Status(http.StatusAccepted).Json(httpcontract.Json{"data": job})
}

func (c *BackupController) Show(ctx httpcontract.Context) httpcontract.Response {
	job, err := c.service.Find(uint(ctx.Request().RouteInt64("id")))
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "BACKUP_NOT_FOUND")
	}
	job.StoragePath = ""
	job.ErrorMessage = safeErrorMessage(job.ErrorMessage)
	return ctx.Response().Success().Json(httpcontract.Json{"data": job})
}

func (c *BackupController) Download(ctx httpcontract.Context) httpcontract.Response {
	job, err := c.service.Find(uint(ctx.Request().RouteInt64("id")))
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "BACKUP_NOT_FOUND")
	}
	filePath, err := c.service.DownloadPath(job)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusConflict, "BACKUP_NOT_READY")
	}
	return ctx.Response().Download(filePath, job.FileName)
}

func (c *BackupController) Delete(ctx httpcontract.Context) httpcontract.Response {
	operatorID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	if err := c.service.Delete(operatorID, uint(ctx.Request().RouteInt64("id"))); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "BACKUP_NOT_FOUND")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": true})
}

func (c *BackupController) Validate(ctx httpcontract.Context) httpcontract.Response {
	operatorID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	filename, content, err := readBackupMultipart(ctx.Request().Origin(), maxBackupUploadBytes)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusBadRequest, "BACKUP_UPLOAD_INVALID")
	}
	job, preview, err := c.service.ValidateFile(ctx.Context(), operatorID, filename, content)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "BACKUP_ARCHIVE_INVALID")
	}
	job.StoragePath = ""
	job.ErrorMessage = ""
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": httpcontract.Json{"job": job, "preview": preview}})
}

func (c *BackupController) Restore(ctx httpcontract.Context) httpcontract.Response {
	var input backupservices.RestoreRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusBadRequest, "BACKUP_RESTORE_INVALID")
	}
	job, err := c.restore.Restore(ctx.Context(), input)
	if errors.Is(err, backupservices.ErrRestoreModeNotAllowed) {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "BACKUP_RESTORE_MODE_INVALID")
	}
	if errors.Is(err, backupservices.ErrRestoreConfirmationRequired) {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "BACKUP_RESTORE_CONFIRMATION_REQUIRED")
	}
	if errors.Is(err, backupservices.ErrRestoreTargetNotEmpty) {
		return adminmiddleware.APIError(ctx, http.StatusConflict, "BACKUP_RESTORE_TARGET_NOT_EMPTY")
	}
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "BACKUP_RESTORE_FAILED")
	}
	job.StoragePath = ""
	job.ErrorMessage = ""
	return ctx.Response().Success().Json(httpcontract.Json{"data": job})
}

func readBackupMultipart(request *http.Request, maxBytes int64) (string, []byte, error) {
	if request == nil || maxBytes <= 0 {
		return "", nil, errors.New("invalid backup upload")
	}
	if request.MultipartForm == nil {
		if err := request.ParseMultipartForm(4 << 20); err != nil {
			return "", nil, err
		}
	}
	defer request.MultipartForm.RemoveAll()
	files := request.MultipartForm.File["file"]
	if len(files) != 1 || files[0].Size > maxBytes {
		return "", nil, errors.New("invalid backup file")
	}
	file, err := files[0].Open()
	if err != nil {
		return "", nil, err
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil || int64(len(body)) > maxBytes || len(body) == 0 {
		return "", nil, errors.New("invalid backup file")
	}
	return files[0].Filename, body, nil
}

func authenticatedUserID(ctx httpcontract.Context) (uint, error) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || value == 0 {
		return 0, errors.New("invalid authenticated user")
	}
	return uint(value), nil
}

func safeErrorMessage(value string) string {
	if strings.Contains(strings.ToLower(value), "password") {
		return "backup job failed"
	}
	return value
}
