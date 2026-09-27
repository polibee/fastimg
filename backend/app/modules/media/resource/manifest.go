package media

import "goravel/app/core/resource"

func Manifest() resource.Manifest {
	return resource.Manifest{
		Name: "media", Label: "Media library", Route: "/admin/media", Table: "media_assets", Permissions: []string{"admin.media.view"},
		Navigation: resource.Navigation{Group: "business", Order: 80}, DataScope: resource.DataScopeAll, SoftDelete: true,
		Fields: []resource.Field{
			{Name: "user_id", Label: "User ID", Type: "integer", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "folder_id", Label: "Folder ID", Type: "integer", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "original_name", Label: "Original name", Type: "text", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "content_type", Label: "Content type", Type: "text", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "format", Label: "Format", Type: "text", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "size_bytes", Label: "Size (bytes)", Type: "integer", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "width", Label: "Width", Type: "integer", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "height", Label: "Height", Type: "integer", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
			{Name: "status", Label: "Status", Type: "select", Required: true, Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true, Options: []resource.Option{{Value: "processing", Label: "Processing"}, {Value: "ready", Label: "Ready"}, {Value: "blocked", Label: "Blocked"}, {Value: "deleted", Label: "Deleted"}}},
			{Name: "visibility", Label: "Visibility", Type: "select", Required: true, Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true, Options: []resource.Option{{Value: "private", Label: "Private"}, {Value: "link", Label: "Unlisted"}, {Value: "public", Label: "Public"}}},
			{Name: "moderation_status", Label: "Moderation status", Type: "select", Required: true, Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true, Options: []resource.Option{{Value: "pending", Label: "Pending"}, {Value: "manual_review", Label: "Manual review"}, {Value: "approved", Label: "Approved"}, {Value: "rejected", Label: "Rejected"}}},
			{Name: "deleted_at", Label: "Deleted at", Type: "datetime-local", Visible: true, Readable: true, Writable: false, Sensitive: false, PolicyConfigured: true},
		},
		Columns: []resource.Column{{Name: "id", Label: "ID", Sortable: true}, {Name: "user_id", Label: "User ID", Sortable: true}, {Name: "original_name", Label: "Original name", Sortable: true}, {Name: "content_type", Label: "Content type", Sortable: true}, {Name: "size_bytes", Label: "Size (bytes)", Sortable: true}, {Name: "status", Label: "Status", Sortable: true}, {Name: "visibility", Label: "Visibility", Sortable: true}, {Name: "moderation_status", Label: "Moderation status", Sortable: true}, {Name: "created_at", Label: "Created at", Sortable: true}},
		Filters: []resource.Filter{{Name: "user_id", Label: "User ID", Type: "select"}, {Name: "status", Label: "Status", Type: "select", Options: []resource.Option{{Value: "processing", Label: "Processing"}, {Value: "ready", Label: "Ready"}, {Value: "blocked", Label: "Blocked"}, {Value: "deleted", Label: "Deleted"}}}, {Name: "visibility", Label: "Visibility", Type: "select", Options: []resource.Option{{Value: "private", Label: "Private"}, {Value: "link", Label: "Unlisted"}, {Value: "public", Label: "Public"}}}, {Name: "moderation_status", Label: "Moderation status", Type: "select", Options: []resource.Option{{Value: "pending", Label: "Pending"}, {Value: "manual_review", Label: "Manual review"}, {Value: "approved", Label: "Approved"}, {Value: "rejected", Label: "Rejected"}}}},
		Actions: []resource.Action{
			{Name: "view", Label: "View", Permission: "admin.media.view"},
			{Name: "moderate", Label: "Moderate media", Permission: "admin.media.update", Batch: true, Kind: "media-moderation", Payload: "media-moderation", PayloadFields: []resource.ActionPayloadField{{Name: "operation", Label: "Operation", Type: "select", Required: true, Options: []resource.Option{{Value: "hide", Label: "Hide"}, {Value: "restore", Label: "Restore"}, {Value: "approve", Label: "Approve"}, {Value: "reject", Label: "Reject"}}}}},
			{Name: "permanent-delete", Label: "Permanently delete", Permission: "admin.media.delete", Batch: true, Kind: "media-permanent-delete", Payload: "media-permanent-delete", PayloadFields: []resource.ActionPayloadField{{Name: "confirmation", Label: "Confirmation", Type: "text", Required: true}}},
		},
	}
}
