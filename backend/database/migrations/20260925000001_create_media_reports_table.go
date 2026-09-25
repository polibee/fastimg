package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260925000001CreateMediaReportsTable stores user reports separately from
// the media asset and from the payment/audit ledgers. A report is a moderation
// queue item; resolving it may update media status but never deletes evidence.
type M20260925000001CreateMediaReportsTable struct{}

func (m *M20260925000001CreateMediaReportsTable) Signature() string {
	return "20260925000001_create_media_reports_table"
}

func (m *M20260925000001CreateMediaReportsTable) Up() error {
	if facades.Schema().HasTable("media_reports") {
		return nil
	}
	return facades.Schema().Create("media_reports", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("media_asset_id")
		table.UnsignedBigInteger("reporter_id").Nullable()
		table.String("reason", 64)
		table.Text("description").Nullable()
		table.String("status", 24).Default("pending")
		table.Text("resolution").Nullable()
		table.UnsignedBigInteger("resolved_by").Nullable()
		table.DateTimeTz("resolved_at").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Index("media_asset_id", "status")
		table.Index("reporter_id", "created_at")
	})
}

func (m *M20260925000001CreateMediaReportsTable) Down() error {
	return facades.Schema().DropIfExists("media_reports")
}
