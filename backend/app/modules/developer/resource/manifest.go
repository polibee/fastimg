package resource

import "goravel/app/core/resource"

// Manifest exposes operational metadata for Personal API Tokens. The raw
// token and digest are intentionally not declared as fields, so the generic
// Resource Engine cannot project or export them.
func Manifest() resource.Manifest {
	return resource.Manifest{
		Name:        "api_tokens",
		Label:       "API Tokens",
		Route:       "/admin/api_tokens",
		Table:       "api_tokens",
		Permissions: []string{"admin.api_tokens.view"},
		Navigation:  resource.Navigation{Group: "developer", Order: 10},
		Fields: []resource.Field{
			{Name: "user_id", Label: "User ID", Type: "integer", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "name", Label: "Name", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "token_prefix", Label: "Token prefix", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "scopes_json", Label: "Scopes", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "status", Label: "Status", Type: "select", Options: []resource.Option{{Value: "active", Label: "Active"}, {Value: "disabled", Label: "Disabled"}, {Value: "revoked", Label: "Revoked"}}, Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "expires_at", Label: "Expires at", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "last_used_at", Label: "Last used at", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "last_used_ip", Label: "Last used IP", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "usage_count", Label: "Usage count", Type: "integer", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "revoked_at", Label: "Revoked at", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
			{Name: "created_at", Label: "Created at", Type: "text", Visible: true, Readable: true, Writable: false, PolicyConfigured: true},
		},
		Columns: []resource.Column{
			{Name: "user_id", Label: "User ID", Sortable: true},
			{Name: "name", Label: "Name", Sortable: true},
			{Name: "token_prefix", Label: "Token prefix", Sortable: false},
			{Name: "scopes_json", Label: "Scopes", Sortable: false},
			{Name: "status", Label: "Status", Sortable: true},
			{Name: "expires_at", Label: "Expires at", Sortable: true},
			{Name: "last_used_at", Label: "Last used at", Sortable: true},
		},
		Actions: []resource.Action{
			{Name: "view", Label: "View", Permission: "admin.api_tokens.view"},
			{Name: "set-status", Label: "Set status", Kind: "api-token-status", Permission: "admin.api_tokens.update", Batch: true, Payload: "api-token-status", PayloadFields: []resource.ActionPayloadField{{Name: "status", Label: "Status", Type: "select", Required: true, Options: []resource.Option{{Value: "disabled", Label: "Disabled"}, {Value: "revoked", Label: "Revoked"}}}}},
			{Name: "delete", Label: "Delete", Permission: "admin.api_tokens.delete"},
		},
		Filters: []resource.Filter{{Name: "status", Label: "Status", Type: "select", Options: []resource.Option{{Value: "active", Label: "Active"}, {Value: "disabled", Label: "Disabled"}, {Value: "revoked", Label: "Revoked"}}}},
	}
}
