package controllers

import (
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	billingservices "goravel/app/services/billing"
)

type componentProbe struct {
	Name   string
	Status string
	Detail string
}

type Controller struct{}

func NewController() *Controller { return &Controller{} }

func (c *Controller) Index(ctx httpcontract.Context) httpcontract.Response {
	probes := []componentProbe{
		{Name: "api", Status: "operational"},
		{Name: "database", Status: "operational"},
		{Name: "storage", Status: "operational"},
		{Name: "payments", Status: "operational"},
	}
	if facades.Schema().HasTable("users") {
		if _, err := facades.Orm().Query().Table("users").Count(); err != nil {
			probes[1].Status = "degraded"
		}
	}
	if facades.Schema().HasTable("storage_connections") {
		if count, err := facades.Orm().Query().Table("storage_connections").Where("enabled = ? AND is_primary = ?", true, true).Count(); err != nil || count == 0 {
			probes[2].Status = "degraded"
		}
	}
	if len(billingservices.DefaultGatewayRegistry().Codes()) == 0 {
		probes[3].Status = "degraded"
	}
	maintenance := []map[string]any{}
	if facades.Schema().HasTable("announcements") {
		var rows []map[string]any
		query := facades.Orm().Query().Table("announcements").Where("status = ?", "published")
		if facades.Schema().HasColumn("announcements", "starts_at") {
			query = query.Where("(starts_at IS NULL OR starts_at <= ?)", time.Now().UTC())
			query = query.Where("(ends_at IS NULL OR ends_at >= ?)", time.Now().UTC())
		}
		if err := query.OrderByDesc("id").Limit(20).Get(&rows); err == nil {
			for _, row := range rows {
				maintenance = append(maintenance, map[string]any{
					"id": row["id"], "title": row["title"], "body": row["body"],
					"severity": row["severity"], "starts_at": row["starts_at"], "ends_at": row["ends_at"],
				})
			}
		}
	}
	status := "operational"
	for _, probe := range probes {
		if probe.Status != "operational" {
			status = "degraded"
			break
		}
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{
		"status": status, "updated_at": time.Now().UTC(), "components": publicComponents(probes), "maintenance": maintenance,
	}})
}

func publicComponents(probes []componentProbe) []map[string]any {
	components := make([]map[string]any, 0, len(probes))
	for _, probe := range probes {
		status := probe.Status
		if status == "" {
			status = "unknown"
		}
		components = append(components, map[string]any{
			"name": probe.Name, "status": status, "detail": probe.Name + " is " + status,
		})
	}
	return components
}
