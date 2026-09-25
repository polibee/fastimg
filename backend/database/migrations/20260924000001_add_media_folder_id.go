package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260924000001AddMediaFolderID links an owned media asset to an owned folder.
// The nullable value deliberately keeps uploads independent from folder setup.
type M20260924000001AddMediaFolderID struct{}

func (m *M20260924000001AddMediaFolderID) Signature() string {
	return "20260924000001_add_media_folder_id"
}

func (m *M20260924000001AddMediaFolderID) Up() error {
	if !facades.Schema().HasTable("media_assets") || facades.Schema().HasColumn("media_assets", "folder_id") {
		return nil
	}
	return facades.Schema().Table("media_assets", func(table schema.Blueprint) {
		table.UnsignedBigInteger("folder_id").Nullable()
		table.Index("user_id", "folder_id", "status", "id")
	})
}

func (m *M20260924000001AddMediaFolderID) Down() error {
	if !facades.Schema().HasTable("media_assets") || !facades.Schema().HasColumn("media_assets", "folder_id") {
		return nil
	}
	return facades.Schema().Table("media_assets", func(table schema.Blueprint) {
		table.DropColumn("folder_id")
	})
}
