package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923000002CreateFastimgMediaTables struct{}

func (m *M20260923000002CreateFastimgMediaTables) Signature() string {
	return "20260923000002_create_fastimg_media_tables"
}

func (m *M20260923000002CreateFastimgMediaTables) Up() error {
	if !facades.Schema().HasTable("storage_objects") {
		if err := facades.Schema().Create("storage_objects", func(table schema.Blueprint) {
			table.ID()
			table.String("provider", 32)
			table.String("object_key", 512)
			table.String("content_type", 120)
			table.BigInteger("size_bytes")
			table.String("sha256", 64)
			table.String("status", 24).Default("pending")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("object_key")
			table.Index("status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("media_assets") {
		if err := facades.Schema().Create("media_assets", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.String("original_name", 255)
			table.String("content_type", 120)
			table.String("format", 16)
			table.BigInteger("size_bytes")
			table.String("sha256", 64)
			table.BigInteger("width")
			table.BigInteger("height")
			table.String("status", 24).Default("processing")
			table.DateTimeTz("deleted_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("user_id", "status", "id")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("media_variants") {
		if err := facades.Schema().Create("media_variants", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("media_asset_id")
			table.UnsignedBigInteger("storage_object_id")
			table.String("name", 24)
			table.BigInteger("width")
			table.BigInteger("height")
			table.String("status", 24).Default("pending")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("media_asset_id", "name")
			table.Index("storage_object_id")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("upload_sessions") {
		if err := facades.Schema().Create("upload_sessions", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.UnsignedBigInteger("media_asset_id")
			table.String("idempotency_key", 160)
			table.String("sha256", 64)
			table.BigInteger("size_bytes")
			table.BigInteger("reserved_bytes").Default(0)
			table.String("status", 24).Default("processing")
			table.String("error_code", 80).Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("user_id", "idempotency_key")
			table.Index("user_id", "status")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *M20260923000002CreateFastimgMediaTables) Down() error {
	for _, table := range []string{"upload_sessions", "media_variants", "media_assets", "storage_objects"} {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}
	return nil
}
