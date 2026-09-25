package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260925000003CreateAPITokenRateLimitsTable struct{}

func (m *M20260925000003CreateAPITokenRateLimitsTable) Signature() string {
	return "20260925000003_create_api_token_rate_limits_table"
}

func (m *M20260925000003CreateAPITokenRateLimitsTable) Up() error {
	if facades.Schema().HasTable("api_token_rate_limits") {
		return nil
	}
	return facades.Schema().Create("api_token_rate_limits", func(table schema.Blueprint) {
		table.ID()
		table.String("key_hash", 64)
		table.UnsignedBigInteger("requests").Default(0)
		table.DateTimeTz("window_started_at")
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("key_hash")
		table.Index("window_started_at")
	})
}

func (m *M20260925000003CreateAPITokenRateLimitsTable) Down() error {
	return facades.Schema().DropIfExists("api_token_rate_limits")
}
