package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type OverviewController struct{}

func NewOverviewController() *OverviewController { return &OverviewController{} }

func (o *OverviewController) Index(ctx http.Context) http.Response {
	users, err := tableCount("users")
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "INTERNAL_ERROR"})
	}
	roles, err := tableCount("roles")
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "INTERNAL_ERROR"})
	}
	permissions, err := tableCount("permissions")
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "INTERNAL_ERROR"})
	}
	return ctx.Response().Success().Json(http.Json{"data": http.Json{
		"users": users, "roles": roles, "permissions": permissions,
		"media": mustTableCount("media_assets"), "albums": mustTableCount("albums"), "folders": mustTableCount("folders"),
		"orders": mustTableCount("orders"), "payment_transactions": mustTableCount("payment_transactions"),
	}})
}

func tableCount(table string) (int64, error) {
	if !facades.Schema().HasTable(table) {
		return 0, nil
	}
	return facades.Orm().Query().Table(table).Count()
}

func mustTableCount(table string) int64 {
	count, _ := tableCount(table)
	return count
}
