package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	"goravel/app/models"
	linkservices "goravel/app/services/links"
	mediaservices "goravel/app/services/media"
	planservices "goravel/app/services/plans"
	storageservices "goravel/app/services/storage"
)

type MediaController struct {
	service *mediaservices.MediaLibraryService
	links   *linkservices.Service
}

type moveMediaFolderRequest struct {
	FolderID *uint `json:"folder_id"`
}

type updateVisibilityRequest struct {
	Visibility string `json:"visibility"`
}

type updateLifecycleRequest struct {
	ExpiresAt *string `json:"expires_at"`
}

func NewMediaController() *MediaController {
	provider := storageservices.NewLocalProvider(facades.Storage().Disk("fastimg"))
	return &MediaController{service: mediaservices.NewMediaLibraryService(mediaservices.NewDatabaseRepository(), provider), links: linkservices.NewService(provider, facades.Config().GetString("app.key", ""))}
}

func (c *MediaController) Index(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	page := max(1, ctx.Request().QueryInt("page", 1))
	perPage := max(1, min(100, ctx.Request().QueryInt("per_page", 24)))
	trash := ctx.Request().Query("state") == "trash"
	items, total, err := c.service.List(ctx.Context(), userID, trash, ctx.Request().Query("search"), page, perPage)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "MEDIA_UNAVAILABLE")
	}
	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data = append(data, mediaListItemJSON(item))
	}
	return ctx.Response().Success().Json(httpcontract.Json{
		"data": data,
		"meta": map[string]any{"page": page, "per_page": perPage, "total": total},
	})
}

func (c *MediaController) Show(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	item, err := c.service.GetDetails(ctx.Context(), userID, uint(mediaID))
	if err != nil {
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": c.mediaDetailJSON(ctx, userID, item)})
}

func (c *MediaController) Content(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	variant := ctx.Request().Query("variant", "original")
	content, err := c.service.GetContent(ctx.Context(), userID, uint(mediaID), variant)
	if err != nil {
		if errors.Is(err, planservices.ErrBandwidthQuotaExceeded) {
			return adminmiddleware.APIError(ctx, http.StatusTooManyRequests, "BANDWIDTH_QUOTA_EXCEEDED")
		}
		if errors.Is(err, planservices.ErrBandwidthUnavailable) {
			return adminmiddleware.APIError(ctx, http.StatusServiceUnavailable, "BANDWIDTH_METERING_UNAVAILABLE")
		}
		if errors.Is(err, mediaservices.ErrMediaNotFound) || errors.Is(err, mediaservices.ErrVariantNotFound) || errors.Is(err, mediaservices.ErrInvalidVariantName) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
		}
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "MEDIA_CONTENT_UNAVAILABLE")
	}
	return ctx.Response().Header("Cache-Control", "private, no-store").
		Header("X-Content-Type-Options", "nosniff").Data(http.StatusOK, content.ContentType, content.Bytes)
}

func (c *MediaController) UpdateVisibility(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	var input updateVisibilityRequest
	if err := ctx.Request().Bind(&input); err != nil || !mediaservices.ValidVisibility(input.Visibility) {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_VISIBILITY_INVALID")
	}
	asset, err := c.service.UpdateVisibility(ctx.Context(), userID, uint(mediaID), input.Visibility)
	if err != nil {
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": asset.ID, "visibility": asset.Visibility}})
}

func (c *MediaController) UpdateLifecycle(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	var input updateLifecycleRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_EXPIRY_INVALID")
	}
	var expiresAt *time.Time
	if input.ExpiresAt != nil && strings.TrimSpace(*input.ExpiresAt) != "" {
		parsed, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*input.ExpiresAt))
		if parseErr != nil {
			return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_EXPIRY_INVALID")
		}
		expiresAt = &parsed
	}
	asset, err := c.service.UpdateExpiry(ctx.Context(), userID, uint(mediaID), expiresAt)
	if err != nil {
		if errors.Is(err, mediaservices.ErrInvalidMediaExpiry) {
			return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_EXPIRY_INVALID")
		}
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": asset.ID, "expires_at": asset.ExpiresAt}})
}

// AdminContent is deliberately separate from the member own-scope endpoint.
// The admin RBAC middleware authorizes this all-user preview path.
func (c *MediaController) AdminContent(ctx httpcontract.Context) httpcontract.Response {
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	variantName := ctx.Request().Query("variant", "original")
	if variantName != "original" {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaID, "ready").First(&asset); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	var variant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", asset.ID, variantName, "ready").First(&variant); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Where("id = ? AND status = ?", variant.StorageObjectID, "ready").First(&object); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	content, err := c.serviceStorage().Get(ctx.Context(), object.ObjectKey)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "MEDIA_CONTENT_UNAVAILABLE")
	}
	return ctx.Response().Header("Cache-Control", "private, no-store").Header("X-Content-Type-Options", "nosniff").Data(http.StatusOK, object.ContentType, content)
}

func (c *MediaController) serviceStorage() storageservices.StorageProvider {
	return storageservices.NewLocalProvider(facades.Storage().Disk("fastimg"))
}

func (c *MediaController) Delete(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	asset, err := c.service.SoftDelete(ctx.Context(), userID, uint(mediaID))
	if err != nil {
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": asset.ID, "status": asset.Status}})
}

func (c *MediaController) PermanentDelete(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	if !mediaservices.PermanentDeleteConfirmed(ctx.Request().Input("confirm")) {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_DELETE_CONFIRMATION_REQUIRED")
	}
	asset, err := c.service.PermanentDelete(ctx.Context(), userID, uint(mediaID))
	if err != nil {
		if errors.Is(err, mediaservices.ErrMediaSharedStorage) {
			return adminmiddleware.APIError(ctx, http.StatusConflict, "MEDIA_SHARED_STORAGE")
		}
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": asset.ID, "status": asset.Status}})
}

func (c *MediaController) EmptyTrash(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	if !mediaservices.EmptyTrashConfirmed(ctx.Request().Input("confirm")) {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_TRASH_CONFIRMATION_REQUIRED")
	}
	deleted, err := c.service.EmptyTrash(ctx.Context(), userID)
	if err != nil {
		if errors.Is(err, mediaservices.ErrMediaSharedStorage) {
			return adminmiddleware.APIError(ctx, http.StatusConflict, "MEDIA_SHARED_STORAGE")
		}
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"deleted_count": deleted}})
}

func (c *MediaController) Restore(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	asset, err := c.service.Restore(ctx.Context(), userID, uint(mediaID))
	if err != nil {
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": asset.ID, "status": asset.Status}})
}

func (c *MediaController) MoveToFolder(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	var input moveMediaFolderRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "MEDIA_FOLDER_VALIDATION_FAILED")
	}
	asset, err := c.service.AssignFolder(ctx.Context(), userID, uint(mediaID), input.FolderID)
	if err != nil {
		if errors.Is(err, mediaservices.ErrFolderNotFound) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "FOLDER_NOT_FOUND")
		}
		return mediaServiceFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": asset.ID, "folder_id": asset.FolderID}})
}

func mediaListItemJSON(item mediaservices.MediaListItem) map[string]any {
	variants := make(map[string]any, len(item.Variants))
	for _, variant := range item.Variants {
		variants[variant.Variant.Name] = map[string]any{
			"url":          mediaContentURL(item.Asset.ID, variant.Variant.Name),
			"content_type": variant.Object.ContentType,
			"size_bytes":   variant.Object.SizeBytes,
			"width":        variant.Variant.Width,
			"height":       variant.Variant.Height,
		}
	}
	var links map[string]string
	if item.Asset.Status == "ready" {
		links = mediaservices.Links(item.Asset.ID, item.Asset.OriginalName)
	}
	return map[string]any{
		"id": item.Asset.ID, "original_name": item.Asset.OriginalName,
		"folder_id":    item.Asset.FolderID,
		"content_type": item.Asset.ContentType, "format": item.Asset.Format,
		"size_bytes": item.Asset.SizeBytes, "width": item.Asset.Width,
		"height": item.Asset.Height, "status": item.Asset.Status,
		"visibility": item.Asset.Visibility, "moderation_status": item.Asset.ModerationStatus,
		"expires_at": item.Asset.ExpiresAt,
		"deleted_at": item.Asset.DeletedAt, "created_at": item.Asset.CreatedAt,
		"links": links, "variants": variants,
	}
}

func (c *MediaController) mediaDetailJSON(ctx httpcontract.Context, userID uint, item mediaservices.MediaListItem) map[string]any {
	result := mediaListItemJSON(item)
	if item.Asset.Status == "ready" {
		if variants, err := c.links.CreateStableURLs(ctx.Context(), userID, item.Asset.ID); err == nil {
			result["links"] = mediaservices.LinkFormatsFromVariants(item.Asset.OriginalName, variants)
		}
	}
	return result
}

func mediaServiceFailure(ctx httpcontract.Context, err error) httpcontract.Response {
	if errors.Is(err, mediaservices.ErrMediaNotFound) {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "MEDIA_OPERATION_FAILED")
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

func mediaContentURL(mediaID uint, variant string) string {
	return "/api/v1/media/" + strconv.FormatUint(uint64(mediaID), 10) + "/content?variant=" + variant
}
