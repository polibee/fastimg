package controllers

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

// MediaAccessController exposes the redacted, append-only delivery decisions
// to administrators. Signatures, share tokens and passwords never live in
// media_access_logs and therefore cannot be returned by this endpoint.
type MediaAccessController struct{}

func NewMediaAccessController() *MediaAccessController { return &MediaAccessController{} }

func (c *MediaAccessController) Index(ctx http.Context) http.Response {
	query := resourceListQuery{
		Page:    positiveInt(ctx.Request().Query("page", "1"), 1),
		PerPage: positiveInt(ctx.Request().Query("per_page", "20"), 20),
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}

	q := facades.Orm().Query().Table("media_access_logs")
	if value := strings.TrimSpace(ctx.Request().Query("media_id")); value != "" {
		q = q.Where("media_asset_id = ?", value)
	}
	if value := strings.TrimSpace(ctx.Request().Query("variant")); value != "" {
		q = q.Where("variant = ?", value)
	}
	if value := strings.TrimSpace(ctx.Request().Query("delivery_mode")); value != "" {
		q = q.Where("delivery_mode = ?", value)
	}
	if value := strings.TrimSpace(ctx.Request().Query("result")); value != "" {
		q = q.Where("result = ?", value)
	}
	if value := strings.TrimSpace(ctx.Request().Query("referer_host")); value != "" {
		q = q.Where("referer_host = ?", value)
	}
	q = q.OrderByDesc("id")

	var entries []map[string]any
	var total int64
	if err := q.Paginate(query.Page, query.PerPage, &entries, &total); err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "INTERNAL_ERROR"})
	}
	return resourceListResponse(ctx, entries, query, total)
}
