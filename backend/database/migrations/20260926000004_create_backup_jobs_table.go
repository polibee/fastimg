package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260926000004CreateBackupJobsTable struct{}

func (m *M20260926000004CreateBackupJobsTable) Signature() string {
	return "20260926000004_create_backup_jobs_table"
}

func (m *M20260926000004CreateBackupJobsTable) Up() error {
	if facades.Schema().HasTable("backup_jobs") {
		return nil
	}
	return facades.Schema().Create("backup_jobs", func(table schema.Blueprint) {
		table.ID()
		table.String("kind", 24).Default("site")
		table.String("status", 32).Default("queued")
		table.String("file_name", 255).Nullable()
		table.String("storage_path", 1024).Nullable()
		table.Text("manifest_json").Nullable()
		table.BigInteger("size_bytes").Default(0)
		table.String("error_code", 64).Nullable()
		table.Text("error_message").Nullable()
		table.UnsignedBigInteger("created_by").Nullable()
		table.DateTimeTz("started_at").Nullable()
		table.DateTimeTz("completed_at").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Index("status", "created_at")
		table.Index("created_by", "created_at")
	})
}

func (m *M20260926000004CreateBackupJobsTable) Down() error {
	return facades.Schema().DropIfExists("backup_jobs")
}

