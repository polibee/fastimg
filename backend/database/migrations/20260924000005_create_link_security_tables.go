package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260924000005CreateLinkSecurityTables adds the M3 delivery-policy and
// access-observation tables. It is registered for the next local migration
// run, but is not executed by the application startup.
type M20260924000005CreateLinkSecurityTables struct{}

func (m *M20260924000005CreateLinkSecurityTables) Signature() string {
	return "20260924000005_create_link_security_tables"
}

func (m *M20260924000005CreateLinkSecurityTables) Up() error {
	if !facades.Schema().HasTable("media_hotlink_policies") {
		if err := facades.Schema().Create("media_hotlink_policies", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.UnsignedBigInteger("media_asset_id")
			table.String("mode", 24).Default("off")
			table.Boolean("allow_no_referer").Default(false)
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("media_asset_id")
			table.Index("user_id", "media_asset_id")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("hotlink_domains") {
		if err := facades.Schema().Create("hotlink_domains", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.String("host", 255)
			table.String("status", 24).Default("active")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("user_id", "host")
			table.Index("user_id", "status")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("media_access_logs") {
		return facades.Schema().Create("media_access_logs", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("media_asset_id")
			table.UnsignedBigInteger("share_link_id").Nullable()
			table.String("variant", 24)
			table.String("delivery_mode", 24)
			table.String("result", 32)
			table.String("referer_host", 255).Nullable()
			table.Text("user_agent").Nullable()
			table.DateTimeTz("accessed_at")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("media_asset_id", "accessed_at")
			table.Index("result", "accessed_at")
		})
	}
	return nil
}

func (m *M20260924000005CreateLinkSecurityTables) Down() error {
	if err := facades.Schema().DropIfExists("media_access_logs"); err != nil {
		return err
	}
	if err := facades.Schema().DropIfExists("hotlink_domains"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("media_hotlink_policies")
}
