package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260925000004CreateAlbumMediaTable stores the own-scope many-to-many
// relation between ready media and member albums. Folder membership remains a
// single nullable column on media_assets; albums are deliberate collections.
type M20260925000004CreateAlbumMediaTable struct{}

func (m *M20260925000004CreateAlbumMediaTable) Signature() string {
	return "20260925000004_create_album_media_table"
}

func (m *M20260925000004CreateAlbumMediaTable) Up() error {
	if facades.Schema().HasTable("album_media") {
		return nil
	}
	return facades.Schema().Create("album_media", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("album_id")
		table.UnsignedBigInteger("media_asset_id")
		table.UnsignedBigInteger("user_id")
		table.Integer("sort_order").Default(0)
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("album_id", "media_asset_id")
		table.Index("user_id", "album_id", "sort_order", "media_asset_id")
		table.Index("user_id", "media_asset_id")
	})
}

func (m *M20260925000004CreateAlbumMediaTable) Down() error {
	return facades.Schema().DropIfExists("album_media")
}
