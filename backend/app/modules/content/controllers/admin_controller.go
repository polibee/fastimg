package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	contentroot "goravel/app/modules/content"
	contentmodels "goravel/app/modules/content/models"
	contentservices "goravel/app/modules/content/services"
	auditservices "goravel/app/services/audit"
)

type AdminController struct{ service *contentservices.Service }

func NewAdminController() *AdminController {
	return &AdminController{service: contentservices.NewService()}
}

func (c *AdminController) Index(ctx httpcontract.Context) httpcontract.Response {
	page, perPage := positiveInt(ctx.Request().Query("page"), 1), positiveInt(ctx.Request().Query("per_page"), 20)
	result, err := c.service.List(ctx.Context(), page, perPage)
	if err != nil {
		return contentError(ctx, http.StatusInternalServerError, "CONTENT_PAGES_UNAVAILABLE")
	}
	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, adminPagePayload(item))
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": items, "meta": map[string]any{"page": result.Page, "per_page": result.PerPage, "total": result.Total}})
}

func (c *AdminController) Show(ctx httpcontract.Context) httpcontract.Response {
	page, err := c.service.Get(ctx.Context(), uint(ctx.Request().RouteInt64("id")))
	if err != nil {
		return contentError(ctx, http.StatusNotFound, "CONTENT_PAGE_NOT_FOUND")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": adminPagePayload(page)})
}

func (c *AdminController) Create(ctx httpcontract.Context) httpcontract.Response {
	return c.save(ctx, 0)
}

func (c *AdminController) Update(ctx httpcontract.Context) httpcontract.Response {
	return c.save(ctx, uint(ctx.Request().RouteInt64("id")))
}

func (c *AdminController) save(ctx httpcontract.Context, id uint) httpcontract.Response {
	operatorID := authenticatedUserID(ctx)
	var input contentroot.PageRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return contentError(ctx, http.StatusUnprocessableEntity, "CONTENT_PAGE_VALIDATION_FAILED")
	}
	page, err := c.service.SaveDraft(ctx.Context(), input.Input(id, operatorID))
	if err != nil {
		if errors.Is(err, contentservices.ErrPageNotFound) {
			return contentError(ctx, http.StatusNotFound, "CONTENT_PAGE_NOT_FOUND")
		}
		return contentError(ctx, http.StatusUnprocessableEntity, "CONTENT_PAGE_VALIDATION_FAILED")
	}
	recordContentAudit(operatorID, "content_pages.save", page.ID)
	return ctx.Response().Success().Json(httpcontract.Json{"data": adminPagePayload(page)})
}

func (c *AdminController) Publish(ctx httpcontract.Context) httpcontract.Response {
	return c.transition(ctx, true)
}

func (c *AdminController) Archive(ctx httpcontract.Context) httpcontract.Response {
	return c.transition(ctx, false)
}

func (c *AdminController) transition(ctx httpcontract.Context, publish bool) httpcontract.Response {
	operatorID := authenticatedUserID(ctx)
	id := uint(ctx.Request().RouteInt64("id"))
	var page contentmodels.SitePage
	var err error
	if publish {
		page, err = c.service.Publish(ctx.Context(), id, operatorID)
	} else {
		page, err = c.service.Archive(ctx.Context(), id, operatorID)
	}
	if err != nil {
		if errors.Is(err, contentservices.ErrPageNotFound) {
			return contentError(ctx, http.StatusNotFound, "CONTENT_PAGE_NOT_FOUND")
		}
		return contentError(ctx, http.StatusUnprocessableEntity, "CONTENT_PAGE_STATUS_INVALID")
	}
	action := "content_pages.archive"
	if publish {
		action = "content_pages.publish"
	}
	recordContentAudit(operatorID, action, id)
	return ctx.Response().Success().Json(httpcontract.Json{"data": page})
}

func adminPagePayload(page contentmodels.SitePage) map[string]any {
	return map[string]any{
		"id": page.ID, "slug": page.Slug, "title": page.Title, "content_json": page.ContentJSON,
		"excerpt": page.Excerpt, "seo_title": page.SEOTitle, "seo_description": page.SEODescription,
		"status": page.Status, "published_at": page.PublishedAt, "created_by": page.CreatedBy, "updated_by": page.UpdatedBy,
	}
}

func authenticatedUserID(ctx httpcontract.Context) uint {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0
	}
	value, _ := strconv.ParseUint(identity, 10, 32)
	return uint(value)
}

func recordContentAudit(userID uint, action string, pageID uint) {
	if userID > 0 {
		_ = auditservices.NewAuditService().Record(userID, action, map[string]any{"page_id": pageID})
	}
}

func positiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func contentError(ctx httpcontract.Context, status int, code string) httpcontract.Response {
	return adminmiddleware.APIError(ctx, status, code)
}
