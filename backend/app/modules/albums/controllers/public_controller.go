package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	collectionservices "goravel/app/services/collections"
	discoveryservices "goravel/app/services/discovery"
)

type PublicController struct {
	collections *collectionservices.Service
	discovery   *discoveryservices.Service
}

func NewPublicController() *PublicController {
	return &PublicController{
		collections: collectionservices.NewService(collectionservices.NewDatabaseRepository()),
		discovery:   discoveryservices.NewDatabaseServiceWithRuntimeStorage(),
	}
}

func (c *PublicController) Show(ctx httpcontract.Context) httpcontract.Response {
	albumID := ctx.Request().RouteInt64("id")
	if albumID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PUBLIC_ALBUM_NOT_FOUND")
	}
	album, err := c.collections.PublicAlbum(ctx.Context(), uint(albumID))
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PUBLIC_ALBUM_NOT_FOUND")
	}
	media := make([]map[string]any, 0, len(album.Media))
	for _, item := range album.Media {
		media = append(media, map[string]any{
			"id":            item.ID,
			"original_name": item.OriginalName,
			"content_type":  item.ContentType,
			"width":         item.Width,
			"height":        item.Height,
			"size_bytes":    item.SizeBytes,
			"created_at":    item.CreatedAt,
			"thumbnail_url": publicAlbumContentURL(album.ID, item.ID, "original"),
			"original_url":  publicAlbumContentURL(album.ID, item.ID, "original"),
		})
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=60").Success().Json(httpcontract.Json{"data": map[string]any{
		"id": album.ID, "name": album.Name, "visibility": album.Visibility, "created_at": album.CreatedAt, "updated_at": album.UpdatedAt, "media": media,
	}})
}

func (c *PublicController) Content(ctx httpcontract.Context) httpcontract.Response {
	albumID := ctx.Request().RouteInt64("id")
	mediaID := ctx.Request().RouteInt64("media_id")
	if albumID <= 0 || mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PUBLIC_ALBUM_MEDIA_NOT_FOUND")
	}
	album, err := c.collections.PublicAlbum(ctx.Context(), uint(albumID))
	if err != nil || !publicAlbumContains(album, uint(mediaID)) {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PUBLIC_ALBUM_MEDIA_NOT_FOUND")
	}
	variant := ctx.Request().Query("variant", "original")
	contentType, content, err := c.discovery.ContentPublic(ctx.Context(), uint(mediaID), variant)
	if err != nil {
		if errors.Is(err, discoveryservices.ErrDiscoveryVariantNotFound) || errors.Is(err, discoveryservices.ErrDiscoveryMediaNotFound) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "PUBLIC_ALBUM_MEDIA_NOT_FOUND")
		}
		if errors.Is(err, discoveryservices.ErrDiscoveryDisabled) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "PUBLIC_ALBUM_MEDIA_NOT_FOUND")
		}
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "PUBLIC_ALBUM_CONTENT_UNAVAILABLE")
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=300").Header("X-Content-Type-Options", "nosniff").Data(http.StatusOK, contentType, content)
}

func publicAlbumContains(album collectionservices.PublicAlbum, mediaID uint) bool {
	for _, item := range album.Media {
		if item.ID == mediaID {
			return true
		}
	}
	return false
}

func publicAlbumContentURL(albumID, mediaID uint, variant string) string {
	base := strings.TrimRight(facades.Config().GetString("app.url", "http://127.0.0.1:53085"), "/")
	return fmt.Sprintf("%s/api/v1/public/albums/%d/media/%d/content?variant=%s", base, albumID, mediaID, variant)
}
