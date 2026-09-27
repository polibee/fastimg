package migrations

import "goravel/app/facades"

// M20260926000001DefaultUploadedMediaApproved separates ordinary media
// availability from the optional public-discovery review workflow. Existing
// uploads that were never submitted to discovery are approved in place; a
// submitted discovery candidate remains pending for administrator review.
type M20260926000001DefaultUploadedMediaApproved struct{}

func (m *M20260926000001DefaultUploadedMediaApproved) Signature() string {
	return "20260926000001_default_uploaded_media_approved"
}

func (m *M20260926000001DefaultUploadedMediaApproved) Up() error {
	if !facades.Schema().HasTable("media_assets") ||
		!facades.Schema().HasColumn("media_assets", "visibility") ||
		!facades.Schema().HasColumn("media_assets", "moderation_status") ||
		!facades.Schema().HasColumn("media_assets", "discovery_submitted_at") {
		return nil
	}
	if _, err := facades.Orm().Query().Exec("ALTER TABLE media_assets ALTER COLUMN visibility SET DEFAULT 'public'"); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec("ALTER TABLE media_assets ALTER COLUMN moderation_status SET DEFAULT 'approved'"); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Table("media_assets").Where(
		"status = ? AND deleted_at IS NULL AND moderation_status = ? AND discovery_submitted_at IS NULL",
		"ready", "pending",
	).Update("moderation_status", "approved")
	return err
}

func (m *M20260926000001DefaultUploadedMediaApproved) Down() error {
	// Do not downgrade moderation state or expose previously private media on rollback.
	return nil
}
