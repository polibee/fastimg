package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"

	adminmiddleware "goravel/app/http/middleware"
	mediaservices "goravel/app/services/media"
)

type ExportController struct {
	service *mediaservices.MediaExportService
}

func NewExportController() *ExportController {
	return &ExportController{service: mediaservices.NewMediaExportService()}
}

func (c *ExportController) Index(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	page := ctx.Request().QueryInt("page", 1)
	perPage := ctx.Request().QueryInt("per_page", 20)
	jobs, total, err := c.service.ListPage(userID, page, perPage)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "MEDIA_EXPORT_UNAVAILABLE")
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
	return ctx.Response().Success().Json(httpcontract.Json{"data": jobs, "meta": map[string]any{"page": page, "per_page": perPage, "total": total, "last_page": lastPage}})
}

func (c *ExportController) Create(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	job, err := c.service.Create(ctx.Context(), userID)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "MEDIA_EXPORT_CREATE_FAILED")
	}
	return ctx.Response().Status(http.StatusAccepted).Json(httpcontract.Json{"data": job})
}

func (c *ExportController) Show(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	id, _ := strconv.ParseUint(ctx.Request().Route("id"), 10, 32)
	job, err := c.service.Find(userID, uint(id))
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_EXPORT_NOT_FOUND")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": job})
}

func (c *ExportController) Download(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	id, _ := strconv.ParseUint(ctx.Request().Route("id"), 10, 32)
	job, err := c.service.Find(userID, uint(id))
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "MEDIA_EXPORT_NOT_FOUND")
	}
	pathName, err := c.service.DownloadPath(job)
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusConflict, "MEDIA_EXPORT_NOT_READY")
	}
	return ctx.Response().Download(pathName, filepath.Base(pathName))
}
