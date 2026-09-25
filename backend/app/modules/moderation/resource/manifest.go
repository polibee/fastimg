package resource

import "goravel/app/core/resource"

func Manifest() resource.Manifest {
	return resource.Manifest{
		Name: "reports", Label: "Media reports", Route: "/admin/reports", Table: "media_reports",
		Permissions: []string{"admin.reports.view"}, DataScope: resource.DataScopeAll,
		Navigation: resource.Navigation{Group: "moderation", Order: 10},
		Fields: []resource.Field{
			{Name: "media_asset_id", Label: "Media ID", Type: "integer", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "reporter_id", Label: "Reporter ID", Type: "integer", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "reason", Label: "Reason", Type: "text", Required: true, Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "description", Label: "Description", Type: "textarea", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "status", Label: "Status", Type: "select", Required: true, Visible: true, Readable: true, Writable: false, PolicyConfigured: true, Options: []resource.Option{{Value: "pending", Label: "Pending"}, {Value: "in_review", Label: "In review"}, {Value: "resolved", Label: "Resolved"}, {Value: "rejected", Label: "Rejected"}}},
			{Name: "resolution", Label: "Resolution", Type: "textarea", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "resolved_by", Label: "Resolved by", Type: "integer", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "resolved_at", Label: "Resolved at", Type: "datetime-local", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
		},
		Columns: []resource.Column{{Name: "id", Label: "ID", Sortable: true}, {Name: "media_asset_id", Label: "Media ID", Sortable: true}, {Name: "reporter_id", Label: "Reporter ID", Sortable: true}, {Name: "reason", Label: "Reason", Sortable: true}, {Name: "status", Label: "Status", Sortable: true}, {Name: "created_at", Label: "Created at", Sortable: true}},
		Actions: []resource.Action{
			{Name: "view", Label: "View", Permission: "admin.reports.view"},
			{Name: "resolve", Label: "Resolve report", Kind: "report-resolve", Permission: "admin.reports.update", Batch: true, Payload: "report-resolve", PayloadFields: []resource.ActionPayloadField{
				{Name: "action", Label: "Resolution action", Type: "select", Required: true, Options: []resource.Option{{Value: "dismiss", Label: "Dismiss report"}, {Value: "hide_media", Label: "Hide media"}, {Value: "restore_media", Label: "Restore media"}}},
				{Name: "resolution", Label: "Resolution note", Type: "textarea", Required: true},
			}},
		},
	}
}
