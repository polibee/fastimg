package resource

import "goravel/app/core/resource"

func Manifest() resource.Manifest {
	return resource.Manifest{
		Name: "plans", Label: "Plans", Route: "/admin/plans", Table: "plans",
		Permissions: []string{"admin.plans.view"},
		Navigation:  resource.Navigation{Group: "business", Order: 110},
		Fields: []resource.Field{
			{Name: "code", Label: "Code", Type: "text", Required: true},
			{Name: "name", Label: "Name", Type: "text", Required: true},
			{Name: "description", Label: "Description", Type: "text"},
			{Name: "price_amount", Label: "Price amount", Type: "number", Required: true},
			{Name: "currency", Label: "Currency", Type: "text", Required: true},
			{Name: "billing_period", Label: "Billing period", Type: "select", Required: true, Options: []resource.Option{{Value: "monthly", Label: "Monthly"}, {Value: "yearly", Label: "Yearly"}}},
			{Name: "entitlements_json", Label: "Entitlements JSON", Type: "textarea", Required: true},
			{Name: "status", Label: "Status", Type: "select", Required: true, Options: []resource.Option{{Value: "active", Label: "Active"}, {Value: "disabled", Label: "Disabled"}}},
			{Name: "sort_order", Label: "Sort order", Type: "number"},
		},
		Columns: []resource.Column{
			{Name: "code", Label: "Code", Sortable: true}, {Name: "name", Label: "Name", Sortable: true},
			{Name: "price_amount", Label: "Price amount", Sortable: true}, {Name: "currency", Label: "Currency", Sortable: true},
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
