package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260928000008CreateMediaExportJobs struct{}

func (m *M20260928000008CreateMediaExportJobs) Signature() string {
	return "20260928000008_create_media_export_jobs"
}

func (m *M20260928000008CreateMediaExportJobs) Up() error {
	if facades.Schema().HasTable("media_export_jobs") {
		return nil
	}
	return facades.Schema().Create("media_export_jobs", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.String("status", 24).Default("queued")
		table.String("file_name", 255).Nullable()
		table.String("storage_path", 1024).Nullable()
		table.BigInteger("size_bytes").Default(0)
		table.BigInteger("media_count").Default(0)
		table.String("error_code", 80).Nullable()
		table.Text("error_message").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Index("user_id", "status", "id")
	})
}

func (m *M20260928000008CreateMediaExportJobs) Down() error {
	return facades.Schema().DropIfExists("media_export_jobs")
}
