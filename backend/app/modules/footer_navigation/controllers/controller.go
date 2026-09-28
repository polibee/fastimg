package controllers

import (
	"fmt"

	httpcontract "github.com/goravel/framework/contracts/http"
	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	footerservices "goravel/app/modules/footer_navigation/services"
	auditservices "goravel/app/services/audit"
)

type Controller struct{ service *footerservices.Service }

func NewController() *Controller { return &Controller{service: footerservices.NewService()} }

func (c *Controller) Public(ctx httpcontract.Context) httpcontract.Response {
	locale := ctx.Request().Query("locale")
	groups, err := c.service.List(locale, true)
	if err != nil {
		return adminmiddleware.APIError(ctx, 500, "FOOTER_NAVIGATION_UNAVAILABLE")
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=300").Success().Json(httpcontract.Json{"data": groups})
}
func (c *Controller) Index(ctx httpcontract.Context) httpcontract.Response {
	groups, err := c.service.List("", false)
	if err != nil {
		return adminmiddleware.APIError(ctx, 500, "FOOTER_NAVIGATION_UNAVAILABLE")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": groups})
}

func (c *Controller) SaveGroup(ctx httpcontract.Context) httpcontract.Response {
	var input footerservices.GroupInput
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, 422, "FOOTER_NAVIGATION_INVALID")
	}
	id := uint(ctx.Request().RouteInt64("id"))
	row, err := c.service.SaveGroup(id, input)
	if err != nil {
		return adminmiddleware.APIError(ctx, 422, "FOOTER_NAVIGATION_INVALID")
	}
	record(ctx, "footer_navigation.group.save", row.ID)
	return ctx.Response().Success().Json(httpcontract.Json{"data": row})
}
func (c *Controller) SaveItem(ctx httpcontract.Context) httpcontract.Response {
	var input footerservices.ItemInput
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, 422, "FOOTER_NAVIGATION_INVALID")
	}
	id := uint(ctx.Request().RouteInt64("id"))
	row, err := c.service.SaveItem(id, input)
	if err != nil {
		return adminmiddleware.APIError(ctx, 422, "FOOTER_NAVIGATION_INVALID")
	}
	record(ctx, "footer_navigation.item.save", row.ID)
	return ctx.Response().Success().Json(httpcontract.Json{"data": row})
}
func (c *Controller) DeleteGroup(ctx httpcontract.Context) httpcontract.Response {
	id := uint(ctx.Request().RouteInt64("id"))
	if err := c.service.DeleteGroup(id); err != nil {
		return adminmiddleware.APIError(ctx, 404, "FOOTER_NAVIGATION_NOT_FOUND")
	}
	record(ctx, "footer_navigation.group.delete", id)
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": id}})
}
func (c *Controller) DeleteItem(ctx httpcontract.Context) httpcontract.Response {
	id := uint(ctx.Request().RouteInt64("id"))
	if err := c.service.DeleteItem(id); err != nil {
		return adminmiddleware.APIError(ctx, 404, "FOOTER_NAVIGATION_NOT_FOUND")
	}
	record(ctx, "footer_navigation.item.delete", id)
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"id": id}})
}

func record(ctx httpcontract.Context, action string, id uint) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return
	}
	var userID uint
	_, _ = fmt.Sscan(identity, &userID)
	if userID > 0 {
		_ = auditservices.NewAuditService().Record(userID, action, map[string]any{"id": id})
	}
}
