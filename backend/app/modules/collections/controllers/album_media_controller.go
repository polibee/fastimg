package controllers

import (
	"errors"
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"

	adminmiddleware "goravel/app/http/middleware"
	collectionservices "goravel/app/services/collections"
)

type MemberAlbumMediaController struct {
	service *collectionservices.Service
}

type albumMediaRequest struct {
	MediaIDs []uint `json:"media_ids"`
}

type albumMediaMoveRequest struct {
	SourceAlbumID uint   `json:"source_album_id"`
	MediaIDs      []uint `json:"media_ids"`
}

func NewMemberAlbumMediaController() *MemberAlbumMediaController {
	return &MemberAlbumMediaController{service: collectionservices.NewService(collectionservices.NewDatabaseRepository())}
}

func (c *MemberAlbumMediaController) Index(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	mediaIDs, err := c.service.ListAlbumMediaIDs(ctx.Context(), userID, uint(albumID))
	if err != nil {
		return albumMediaError(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{
		"data": mediaIDs,
		"meta": map[string]any{"album_id": albumID, "total": len(mediaIDs)},
	})
}

func (c *MemberAlbumMediaController) Add(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	var input albumMediaRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	}
	mutation, err := c.service.AddMediaToAlbum(ctx.Context(), userID, uint(albumID), input.MediaIDs)
	if err != nil {
		return albumMediaError(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": mutation})
}

func (c *MemberAlbumMediaController) Remove(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	albumID := ctx.Request().RouteInt64("id")
	mediaID := ctx.Request().RouteInt64("media_id")
	if albumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	}
	if err := c.service.RemoveMediaFromAlbum(ctx.Context(), userID, uint(albumID), uint(mediaID)); err != nil {
		return albumMediaError(ctx, err)
	}
	return ctx.Response().NoContent(http.StatusNoContent)
}

func (c *MemberAlbumMediaController) RemoveBatch(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	var input albumMediaRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	}
	mutation, err := c.service.RemoveMediaFromAlbumBatch(ctx.Context(), userID, uint(albumID), input.MediaIDs)
	if err != nil {
		return albumMediaError(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": mutation})
}

func (c *MemberAlbumMediaController) Reorder(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	var input albumMediaRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	}
	if err := c.service.ReorderAlbumMedia(ctx.Context(), userID, uint(albumID), input.MediaIDs); err != nil {
		return albumMediaError(ctx, err)
	}
	ordered, err := c.service.ListAlbumMediaIDs(ctx.Context(), userID, uint(albumID))
	if err != nil {
		return albumMediaError(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": ordered, "meta": map[string]any{"album_id": albumID, "total": len(ordered)}})
}

func (c *MemberAlbumMediaController) Move(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	destinationAlbumID := ctx.Request().RouteInt64("id")
	if destinationAlbumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	}
	var input albumMediaMoveRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	}
	mutation, err := c.service.MoveMediaBetweenAlbums(ctx.Context(), userID, input.SourceAlbumID, uint(destinationAlbumID), input.MediaIDs)
	if err != nil {
		return albumMediaError(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": mutation})
}

func albumMediaError(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, collectionservices.ErrNotFound):
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "ALBUM_NOT_FOUND")
	case errors.Is(err, collectionservices.ErrMediaNotFound):
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_NOT_FOUND")
	case errors.Is(err, collectionservices.ErrInvalidMediaIDs):
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_VALIDATION_FAILED")
	case errors.Is(err, collectionservices.ErrAlbumMediaBatchTooLarge):
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_BATCH_TOO_LARGE")
	case errors.Is(err, collectionservices.ErrAlbumMediaOrderInvalid):
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_ORDER_INVALID")
	case errors.Is(err, collectionservices.ErrAlbumMoveInvalid):
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "ALBUM_MEDIA_MOVE_INVALID")
	default:
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "ALBUM_MEDIA_OPERATION_FAILED")
	}
}
