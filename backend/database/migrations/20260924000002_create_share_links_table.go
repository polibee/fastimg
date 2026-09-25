package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260924000002CreateShareLinksTable struct{}

func (m *M20260924000002CreateShareLinksTable) Signature() string {
	return "20260924000002_create_share_links_table"
}

func (m *M20260924000002CreateShareLinksTable) Up() error {
	if facades.Schema().HasTable("share_links") {
		return nil
	}
	return facades.Schema().Create("share_links", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.UnsignedBigInteger("media_asset_id")
		table.String("token_hash", 64)
		table.String("token_prefix", 16)
		table.String("password_hash", 255).Nullable()
		table.String("status", 24).Default("active")
		table.DateTimeTz("expires_at").Nullable()
		table.DateTimeTz("revoked_at").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("token_hash")
		table.Index("user_id", "status", "id")
		table.Index("media_asset_id", "status")
	})
}

func (m *M20260924000002CreateShareLinksTable) Down() error {
	return facades.Schema().DropIfExists("share_links")
}
