package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260925000005AddDiscoveryStateToMedia keeps public discovery state on the
// media asset. Public discovery reads additionally require an explicit
// discovery history; ordinary public links do not wait for moderation.
type M20260925000005AddDiscoveryStateToMedia struct{}

func (m *M20260925000005AddDiscoveryStateToMedia) Signature() string {
	return "20260925000005_add_discovery_state_to_media"
}

func (m *M20260925000005AddDiscoveryStateToMedia) Up() error {
	if !facades.Schema().HasTable("media_assets") {
		return nil
	}
	if !facades.Schema().HasColumn("media_assets", "visibility") {
		if err := facades.Schema().Table("media_assets", func(table schema.Blueprint) {
			table.String("visibility", 16).Default("public")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("media_assets", "moderation_status") {
		if err := facades.Schema().Table("media_assets", func(table schema.Blueprint) {
			table.String("moderation_status", 24).Default("approved")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("media_assets", "discovery_submitted_at") {
		if err := facades.Schema().Table("media_assets", func(table schema.Blueprint) {
			table.DateTimeTz("discovery_submitted_at").Nullable()
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *M20260925000005AddDiscoveryStateToMedia) Down() error {
	if !facades.Schema().HasTable("media_assets") {
		return nil
	}
	return facades.Schema().Table("media_assets", func(table schema.Blueprint) {
		for _, column := range []string{"discovery_submitted_at", "moderation_status", "visibility"} {
			if facades.Schema().HasColumn("media_assets", column) {
				table.DropColumn(column)
			}
		}
	})
}
