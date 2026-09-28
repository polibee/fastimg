package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260928000005CreateMediaSecurityScans struct{}

func (m *M20260928000005CreateMediaSecurityScans) Signature() string {
	return "20260928000005_create_media_security_scans"
}

func (m *M20260928000005CreateMediaSecurityScans) Up() error {
	if facades.Schema().HasTable("media_security_scans") {
		return nil
	}
	return facades.Schema().Create("media_security_scans", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("media_asset_id")
		table.String("status", 24).Default("queued")
		table.String("provider", 64).Default("local")
		table.String("reason", 120).Nullable()
		table.Text("findings_json").Nullable()
		table.Integer("risk_score").Default(0)
		table.DateTimeTz("started_at").Nullable()
		table.DateTimeTz("completed_at").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("media_asset_id")
		table.Index("status", "created_at")
	})
}

func (m *M20260928000005CreateMediaSecurityScans) Down() error {
	return facades.Schema().DropIfExists("media_security_scans")
}
