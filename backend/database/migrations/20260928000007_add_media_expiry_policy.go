package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260928000007AddMediaExpiryPolicy struct{}

func (m *M20260928000007AddMediaExpiryPolicy) Signature() string {
	return "20260928000007_add_media_expiry_policy"
}

func (m *M20260928000007AddMediaExpiryPolicy) Up() error {
	if !facades.Schema().HasTable("media_assets") {
		return nil
	}
	if !facades.Schema().HasColumn("media_assets", "expires_at") {
		if err := facades.Schema().Table("media_assets", func(table schema.Blueprint) { table.DateTimeTz("expires_at").Nullable() }); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("media_assets", "expiry_notified_at") {
		if err := facades.Schema().Table("media_assets", func(table schema.Blueprint) { table.DateTimeTz("expiry_notified_at").Nullable() }); err != nil {
			return err
		}
	}
	return facades.Schema().Table("media_assets", func(table schema.Blueprint) { table.Index("expires_at", "status") })
}

func (m *M20260928000007AddMediaExpiryPolicy) Down() error { return nil }
