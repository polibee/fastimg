package controllers

import (
	"errors"
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"

	adminmiddleware "goravel/app/http/middleware"
	contentroot "goravel/app/modules/content"
	contentservices "goravel/app/modules/content/services"
)

type PublicController struct{ service *contentservices.Service }

func NewPublicController() *PublicController {
	return &PublicController{service: contentservices.NewService()}
}

func (c *PublicController) Show(ctx httpcontract.Context) httpcontract.Response {
	page, err := c.service.GetPublished(ctx.Context(), ctx.Request().Route("slug"))
	if err != nil {
		if errors.Is(err, contentservices.ErrPageNotFound) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "CONTENT_PAGE_NOT_FOUND")
		}
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "CONTENT_PAGE_UNAVAILABLE")
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=60").Success().Json(httpcontract.Json{"data": contentroot.PublicPagePayload(page)})
}
