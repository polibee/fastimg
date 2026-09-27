package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	linkservices "goravel/app/services/links"
	"goravel/app/services/media"
	planservices "goravel/app/services/plans"
	"goravel/app/services/quota"
	rbacservices "goravel/app/services/rbac"
	settingsservices "goravel/app/services/settings"
	storageservices "goravel/app/services/storage"
)

const (
	maxUploadFileBytes  int64 = 10_000_000
	maxBatchUploadFiles       = 5
	maxBatchUploadBytes int64 = 50_000_000
)

type UploadController struct {
	service *media.UploadService
	links   *linkservices.Service
}

func NewUploadController() *UploadController {
	disk := facades.Storage().Disk("fastimg")
	var provider storageservices.StorageProvider = storageservices.NewLocalProvider(disk)
	storageConnectionID := uint(0)
	if configured, connection, err := storageservices.NewRuntimeRegistry(disk).Primary(); err == nil {
		provider = configured
		storageConnectionID = connection.ID
	}
	return &UploadController{service: media.NewUploadServiceWithConnection(
		media.NewImageProcessor(media.ImageLimits{}), provider, media.NewDatabaseRepository(), storageConnectionID,
	), links: linkservices.NewService(provider, facades.Config().GetString("app.key", ""))}
}

func (c *UploadController) Create(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	upload, err := readMultipartUpload(ctx.Request().Origin(), maxUploadFileBytes)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return uploadFailure(ctx, http.StatusRequestEntityTooLarge, "UPLOAD_REQUEST_TOO_LARGE")
		}
		switch err {
		case errUploadFileRequired:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_FILE_REQUIRED")
		case errMultipleUploadFiles:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_SINGLE_FILE_ONLY")
		case errUploadFileTooLarge:
			return uploadFailure(ctx, http.StatusRequestEntityTooLarge, "UPLOAD_FILE_TOO_LARGE")
		case errUploadFileEmpty:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_FILE_EMPTY")
		default:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_MULTIPART_INVALID")
		}
	}
	watermarkEnabled, watermarkText, watermarkDomain, err := c.watermarkOptions(userID)
	if err != nil {
		facades.Log().Errorf("watermark entitlement unavailable user_id=%d error=%v", userID, err)
		return uploadFailure(ctx, http.StatusConflict, "SUBSCRIPTION_UNAVAILABLE")
	}
	outcome, err := c.service.Upload(ctx.Context(), media.UploadInput{
		UserID: userID, Channel: uploadChannel(ctx), OriginalName: upload.OriginalName,
		DeclaredContentType: upload.ContentType, IdempotencyKey: ctx.Request().Header("Idempotency-Key"),
		Content: upload.Content, WatermarkEnabled: watermarkEnabled, WatermarkText: watermarkText, WatermarkDomain: watermarkDomain,
	})
	if err != nil {
		facades.Log().Errorf("media upload failed user_id=%d error=%v", userID, err)
		status, code := uploadErrorResponse(err)
		return uploadFailure(ctx, status, code)
	}
	responseStatus := http.StatusCreated
	if outcome.Status != "ready" {
		responseStatus = http.StatusAccepted
	}
	var links any
	if outcome.Status == "ready" {
		links = c.readyLinks(ctx, userID, outcome.MediaID, outcome.Name)
	}
	return ctx.Response().Status(responseStatus).Json(httpcontract.Json{"data": httpcontract.Json{
		"id": outcome.MediaID, "upload_session_id": outcome.SessionID, "status": outcome.Status,
		"status_url":    uploadStatusURL(outcome.SessionID),
		"original_name": outcome.Name, "content_type": outcome.ContentType,
		"size_bytes": outcome.SizeBytes, "width": outcome.Width, "height": outcome.Height,
		"replayed": outcome.Replayed, "links": links,
	}})
}

func (c *UploadController) Batch(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	baseKey := strings.TrimSpace(ctx.Request().Header("Idempotency-Key"))
	if len(baseKey) < 8 || len(baseKey) > 160 {
		return uploadFailure(ctx, http.StatusBadRequest, "VALIDATION_ERROR")
	}
	uploads, err := readMultipartUploads(ctx.Request().Origin(), maxUploadFileBytes, maxBatchUploadFiles, maxBatchUploadBytes)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return uploadFailure(ctx, http.StatusRequestEntityTooLarge, "UPLOAD_REQUEST_TOO_LARGE")
		}
		switch err {
		case errUploadFileRequired:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_FILE_REQUIRED")
		case errUploadFileTooLarge:
			return uploadFailure(ctx, http.StatusRequestEntityTooLarge, "UPLOAD_FILE_TOO_LARGE")
		case errBatchFilesTooMany:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_BATCH_TOO_MANY_FILES")
		case errBatchTotalTooLarge:
			return uploadFailure(ctx, http.StatusRequestEntityTooLarge, "UPLOAD_BATCH_TOO_LARGE")
		case errUploadFileEmpty:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_FILE_EMPTY")
		default:
			return uploadFailure(ctx, http.StatusBadRequest, "UPLOAD_MULTIPART_INVALID")
		}
	}
	watermarkEnabled, watermarkText, watermarkDomain, err := c.watermarkOptions(userID)
	if err != nil {
		facades.Log().Errorf("watermark entitlement unavailable user_id=%d error=%v", userID, err)
		return uploadFailure(ctx, http.StatusConflict, "SUBSCRIPTION_UNAVAILABLE")
	}
	items := make([]map[string]any, 0, len(uploads))
	accepted := 0
	failed := 0
	for index, upload := range uploads {
		outcome, uploadErr := c.service.Upload(ctx.Context(), media.UploadInput{
			UserID: userID, Channel: media.UploadChannelMember, OriginalName: upload.OriginalName,
			DeclaredContentType: upload.ContentType, IdempotencyKey: batchIdempotencyKey(baseKey, index),
			Content: upload.Content, WatermarkEnabled: watermarkEnabled, WatermarkText: watermarkText, WatermarkDomain: watermarkDomain,
		})
		if uploadErr != nil {
			_, code := uploadErrorResponse(uploadErr)
			items = append(items, map[string]any{
				"client_name": upload.OriginalName,
				"status":      "failed",
				"error":       map[string]any{"code": code},
			})
			failed++
			continue
		}
		item := c.uploadResultData(ctx, userID, outcome)
		item["client_name"] = upload.OriginalName
		items = append(items, item)
		accepted++
	}
	return ctx.Response().Status(http.StatusMultiStatus).Json(httpcontract.Json{"data": httpcontract.Json{
		"items": items,
		"summary": httpcontract.Json{
			"total": len(items), "accepted": accepted, "failed": failed,
		},
	}})
}

func (c *UploadController) watermarkOptions(userID uint) (bool, string, string, error) {
	administrator, err := rbacservices.NewRBACService().IsAdministrator(userID)
	if err != nil {
		return false, "", "", err
	}
	subscription, err := planservices.NewPlanService().SubscriptionForUser(userID)
	var entitlement quota.Entitlement
	if err == nil {
		entitlement, err = quota.ParseEntitlementJSON(subscription.EntitlementSnapshotJSON)
	}
	watermarkEnabled, entitlementErr := ResolveWatermarkEntitlement(administrator, entitlement, err)
	if entitlementErr != nil {
		return false, "", "", entitlementErr
	}
	settings := settingsservices.NewSettingService()
	watermarkText := settings.Resolve("watermark.text", "FastImg")
	watermarkDomain := settings.Resolve("watermark.domain", "")
	return watermarkEnabled, watermarkText, watermarkDomain, nil
}

func ResolveWatermarkEntitlement(administrator bool, entitlement quota.Entitlement, err error) (bool, error) {
	if err != nil {
		if administrator {
			return false, nil
		}
		return false, err
	}
	return entitlement.WatermarkEnabled, nil
}

func uploadChannel(ctx httpcontract.Context) media.UploadChannel {
	if adminmiddleware.IsMemberTokenRequest(ctx) {
		return media.UploadChannelAPI
	}
	return media.UploadChannelMember
}

func (c *UploadController) Status(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	sessionID := ctx.Request().RouteInt64("id")
	if sessionID <= 0 {
		return uploadFailure(ctx, http.StatusNotFound, "UPLOAD_SESSION_NOT_FOUND")
	}
	outcome, err := c.service.GetStatus(ctx.Context(), userID, uint(sessionID))
	if errors.Is(err, media.ErrMediaNotFound) {
		return uploadFailure(ctx, http.StatusNotFound, "UPLOAD_SESSION_NOT_FOUND")
	}
	if err != nil {
		return uploadFailure(ctx, http.StatusInternalServerError, "UPLOAD_STATUS_UNAVAILABLE")
	}
	status := http.StatusOK
	if outcome.Status == "processing" {
		status = http.StatusAccepted
	}
	var links any
	if outcome.Status == "ready" {
		links = c.readyLinks(ctx, userID, outcome.MediaID, outcome.Name)
	}
	return ctx.Response().Status(status).Json(httpcontract.Json{"data": httpcontract.Json{
		"id": outcome.MediaID, "upload_session_id": outcome.SessionID, "status": outcome.Status,
		"status_url": uploadStatusURL(outcome.SessionID), "replayed": false,
		"original_name": outcome.Name, "content_type": outcome.ContentType,
		"size_bytes": outcome.SizeBytes, "width": outcome.Width, "height": outcome.Height,
		"error_code": outcome.ErrorCode, "links": links,
	}})
}

func (c *UploadController) Retry(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	sessionID := ctx.Request().RouteInt64("id")
	if sessionID <= 0 {
		return uploadFailure(ctx, http.StatusNotFound, "UPLOAD_SESSION_NOT_FOUND")
	}
	outcome, err := c.service.Retry(ctx.Context(), userID, uint(sessionID))
	if err != nil {
		if errors.Is(err, media.ErrMediaNotFound) {
			return uploadFailure(ctx, http.StatusNotFound, "UPLOAD_SESSION_NOT_FOUND")
		}
		status, code := uploadErrorResponse(err)
		return uploadFailure(ctx, status, code)
	}
	responseStatus := http.StatusOK
	if outcome.Status == "processing" {
		responseStatus = http.StatusAccepted
	}
	var links any
	if outcome.Status == "ready" {
		links = c.readyLinks(ctx, userID, outcome.MediaID, outcome.Name)
	}
	return ctx.Response().Status(responseStatus).Json(httpcontract.Json{"data": httpcontract.Json{
		"id": outcome.MediaID, "upload_session_id": outcome.SessionID, "status": outcome.Status,
		"status_url": uploadStatusURL(outcome.SessionID), "replayed": false,
		"original_name": outcome.Name, "content_type": outcome.ContentType,
		"size_bytes": outcome.SizeBytes, "width": outcome.Width, "height": outcome.Height,
		"error_code": outcome.ErrorCode, "links": links,
	}})
}

func uploadErrorResponse(err error) (int, string) {
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		return http.StatusRequestEntityTooLarge, "UPLOAD_REQUEST_TOO_LARGE"
	case errors.Is(err, media.ErrInvalidUploadInput):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	case errors.Is(err, media.ErrImageTooLarge), errors.Is(err, quota.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge, "UPLOAD_FILE_TOO_LARGE"
	case errors.Is(err, quota.ErrStorageQuotaExceeded):
		return http.StatusInsufficientStorage, "STORAGE_QUOTA_EXCEEDED"
	case errors.Is(err, media.ErrDailyUploadLimit):
		return http.StatusTooManyRequests, "DAILY_UPLOAD_LIMIT_REACHED"
	case errors.Is(err, media.ErrMonthlyAPIUploadLimit):
		return http.StatusTooManyRequests, "MONTHLY_API_UPLOAD_LIMIT_REACHED"
	case errors.Is(err, media.ErrMonthlyTransformLimit):
		return http.StatusTooManyRequests, "MONTHLY_TRANSFORM_LIMIT_REACHED"
	case errors.Is(err, media.ErrImageFormatMismatch), errors.Is(err, media.ErrInvalidImage), errors.Is(err, media.ErrUnsupportedImage),
		errors.Is(err, media.ErrImageDimensionsExceeded), errors.Is(err, media.ErrTooManyFrames), errors.Is(err, media.ErrAnimationTooLarge):
		return http.StatusUnprocessableEntity, "IMAGE_FORMAT_INVALID"
	case errors.Is(err, media.ErrIdempotencyConflict):
		return http.StatusConflict, "IDEMPOTENCY_KEY_REUSED"
	case errors.Is(err, media.ErrUploadInProgress):
		return http.StatusConflict, "UPLOAD_IN_PROGRESS"
	case errors.Is(err, media.ErrAccountUnavailable):
		return http.StatusForbidden, "ACCOUNT_UPLOAD_DISABLED"
	case errors.Is(err, media.ErrSubscriptionUnavailable):
		return http.StatusConflict, "SUBSCRIPTION_UNAVAILABLE"
	case errors.Is(err, media.ErrStorageWriteFailed):
		return http.StatusServiceUnavailable, "STORAGE_UNAVAILABLE"
	default:
		return http.StatusInternalServerError, "UPLOAD_FAILED"
	}
}

func uploadFailure(ctx httpcontract.Context, status int, code string) httpcontract.Response {
	return adminmiddleware.APIError(ctx, status, code)
}

func (c *UploadController) uploadResultData(ctx httpcontract.Context, userID uint, outcome media.UploadOutcome) map[string]any {
	var links any
	if outcome.Status == "ready" {
		links = c.readyLinks(ctx, userID, outcome.MediaID, outcome.Name)
	}
	return map[string]any{
		"id": outcome.MediaID, "upload_session_id": outcome.SessionID, "status": outcome.Status,
		"status_url": uploadStatusURL(outcome.SessionID), "original_name": outcome.Name,
		"content_type": outcome.ContentType, "size_bytes": outcome.SizeBytes,
		"width": outcome.Width, "height": outcome.Height, "replayed": outcome.Replayed, "links": links,
	}
}

func (c *UploadController) readyLinks(ctx httpcontract.Context, userID, mediaID uint, originalName string) map[string]string {
	variants, err := c.links.CreateStableURLs(ctx.Context(), userID, mediaID)
	if err != nil {
		// Upload success must not be rolled back because link presentation is
		// unavailable. The fallback remains absolute but is intentionally logged
		// so APP_KEY/APP_URL misconfiguration is visible during development.
		facades.Log().Errorf("public media links unavailable user_id=%d media_id=%d error=%v", userID, mediaID, err)
		return media.Links(mediaID, originalName)
	}
	return media.LinkFormatsFromVariants(originalName, variants)
}

func batchIdempotencyKey(base string, index int) string {
	digest := sha256.Sum256([]byte(base + ":" + strconv.Itoa(index)))
	return "batch-" + hex.EncodeToString(digest[:])
}

func authenticatedUserID(ctx httpcontract.Context) (uint, error) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid authenticated user")
	}
	return uint(parsed), nil
}

func uploadStatusURL(sessionID uint) string {
	return "/api/v1/uploads/" + strconv.FormatUint(uint64(sessionID), 10)
}
