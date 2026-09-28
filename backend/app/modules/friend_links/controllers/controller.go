package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"
	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	models "goravel/app/modules/friend_links/models"
	services "goravel/app/modules/friend_links/services"
	auditservices "goravel/app/services/audit"
)

type Controller struct{ service *services.Service }

func NewController() *Controller { return &Controller{service: services.NewService()} }
func publicPayload(row models.Submission) map[string]any {
	return map[string]any{"id": row.ID, "site_name": row.SiteName, "url": row.URL, "logo_url": row.LogoURL, "description": row.Description}
}

func (c *Controller) PublicList(ctx httpcontract.Context) httpcontract.Response {
	rows, err := c.service.List(models.StatusApproved)
	if err != nil {
		return adminmiddleware.APIError(ctx, 500, "FRIEND_LINKS_UNAVAILABLE")
	}
	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, publicPayload(row))
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=300").Success().Json(httpcontract.Json{"data": data})
}
func (c *Controller) Submit(ctx httpcontract.Context) httpcontract.Response {
	var input services.SubmissionInput
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_INVALID")
	}
	var userID *uint
	if identity, err := facades.Auth(ctx).ID(); err == nil {
		value, _ := strconv.ParseUint(identity, 10, 32)
		if value > 0 {
			id := uint(value)
			userID = &id
		}
	}
	row, err := c.service.Submit(input, userID)
	if err != nil {
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_INVALID")
	}
	return ctx.Response().Status(http.StatusAccepted).Json(httpcontract.Json{"data": map[string]any{"id": row.ID, "status": row.Status}})
}
func (c *Controller) AdminList(ctx httpcontract.Context) httpcontract.Response {
	rows, err := c.service.List(ctx.Request().Query("status"))
	if err != nil {
		return adminmiddleware.APIError(ctx, 500, "FRIEND_LINKS_UNAVAILABLE")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": rows})
}
func (c *Controller) Review(ctx httpcontract.Context) httpcontract.Response {
	var input services.ReviewInput
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_REVIEW_INVALID")
	}
	operatorID := authenticatedID(ctx)
	row, err := c.service.Review(uint(ctx.Request().RouteInt64("id")), input, operatorID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return adminmiddleware.APIError(ctx, 404, "FRIEND_LINK_NOT_FOUND")
		}
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_REVIEW_INVALID")
	}
	if operatorID > 0 {
		_ = auditservices.NewAuditService().Record(operatorID, "friend_links.review", map[string]any{"id": row.ID, "status": row.Status})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": row})
}
func authenticatedID(ctx httpcontract.Context) uint {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0
	}
	var id uint
	_, _ = fmt.Sscan(identity, &id)
	return id
}
