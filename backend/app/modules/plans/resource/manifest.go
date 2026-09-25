package resource

import "goravel/app/core/resource"
import planservices "goravel/app/services/plans"

func Manifest() resource.Manifest {
	return resource.Manifest{
		Name: "plans", Label: "Plans", Route: "/admin/plans", Table: "plans",
		Permissions:   []string{"admin.plans.view"},
		Navigation:    resource.Navigation{Group: "business", Order: 110},
		WritePreparer: planservices.PreparePlanWrite,
		Fields: []resource.Field{
			{Name: "code", Label: "Code", Type: "text", Required: true},
			{Name: "name", Label: "Name", Type: "text", Required: true},
			{Name: "description", Label: "Description", Type: "text"},
			{Name: "entitlements_json", Label: "Entitlements", Type: "entitlements", Hint: "Set storage, upload, API, bandwidth, advertising, and watermark limits using the guided form.", Required: true},
			{Name: "status", Label: "Status", Type: "select", Required: true, Options: []resource.Option{{Value: "active", Label: "Active"}, {Value: "disabled", Label: "Disabled"}}},
			{Name: "sort_order", Label: "Sort order", Type: "number"},
		},
		Columns: []resource.Column{
			{Name: "code", Label: "Code", Sortable: true}, {Name: "name", Label: "Name", Sortable: true},
			{Name: "status", Label: "Status", Sortable: true},
		},
		Actions: []resource.Action{
			{Name: "view", Label: "View", Permission: "admin.plans.view"},
			{Name: "create", Label: "Create", Permission: "admin.plans.create"},
			{Name: "update", Label: "Update", Permission: "admin.plans.update"},
			{Name: "delete", Label: "Delete", Permission: "admin.plans.delete"},
		},
	}
}
