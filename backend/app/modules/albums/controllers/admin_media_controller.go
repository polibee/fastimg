package controllers

import (
	"errors"
	"net/http"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	auditservices "goravel/app/services/audit"
	collectionservices "goravel/app/services/collections"
)

type AdminMediaController struct{ service *collectionservices.Service }

type adminMediaRequest struct {
	MediaIDs []uint `json:"media_ids"`
}

func NewAdminMediaController() *AdminMediaController {
	return &AdminMediaController{service: collectionservices.NewService(collectionservices.NewDatabaseRepository())}
}

func (c *AdminMediaController) Index(ctx httpcontract.Context) httpcontract.Response {
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return albumMediaFailure(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	page := ctx.Request().QueryInt("page", 1)
	perPage := ctx.Request().QueryInt("per_page", 20)
	items, total, err := c.service.ListAdminAlbumMediaPage(ctx.Context(), uint(albumID), page, perPage)
	if err != nil {
		return albumMediaError(ctx, err)
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	lastPage := int64(0)
	if total > 0 {
		lastPage = (total + int64(perPage) - 1) / int64(perPage)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": items, "meta": map[string]any{"album_id": albumID, "page": page, "per_page": perPage, "total": total, "last_page": lastPage}})
}

func (c *AdminMediaController) Add(ctx httpcontract.Context) httpcontract.Response {
	return c.mutate(ctx, "add")
}
func (c *AdminMediaController) Remove(ctx httpcontract.Context) httpcontract.Response {
	return c.mutate(ctx, "remove")
}

func (c *AdminMediaController) mutate(ctx httpcontract.Context, operation string) httpcontract.Response {
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return albumMediaFailure(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	var input adminMediaRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return albumMediaFailure(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	}
	var result collectionservices.AdminAlbumMediaMutation
	var err error
	if operation == "add" {
		result, err = c.service.AddAdminMediaToAlbum(ctx.Context(), uint(albumID), input.MediaIDs)
	} else {
		result, err = c.service.RemoveAdminMediaFromAlbum(ctx.Context(), uint(albumID), input.MediaIDs)
	}
	if err != nil {
		return albumMediaError(ctx, err)
	}
	identity, identityErr := facades.Auth(ctx).ID()
	operatorID, _ := strconv.ParseUint(identity, 10, 32)
	if identityErr == nil && operatorID > 0 {
		_ = auditservices.NewAuditService().Record(uint(operatorID), "admin.albums.media."+operation, map[string]any{"album_id": albumID, "media_ids": result.Changed})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": result})
}

func albumMediaFailure(ctx httpcontract.Context, status int, code string) httpcontract.Response {
	return ctx.Response().Status(status).Json(httpcontract.Json{"code": code})
}

func albumMediaError(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, collectionservices.ErrNotFound):
		return albumMediaFailure(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	case errors.Is(err, collectionservices.ErrMediaNotFound):
		return albumMediaFailure(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	case errors.Is(err, collectionservices.ErrInvalidMediaIDs):
		return albumMediaFailure(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	case errors.Is(err, collectionservices.ErrAlbumMediaBatchTooLarge):
		return albumMediaFailure(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_BATCH_TOO_LARGE")
	default:
		return albumMediaFailure(ctx, http.StatusInternalServerError, "ALBUM_MEDIA_OPERATION_FAILED")
	}
}
