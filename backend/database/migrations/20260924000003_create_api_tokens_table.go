package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260924000003CreateAPITokensTable struct{}

func (m *M20260924000003CreateAPITokensTable) Signature() string {
	return "20260924000003_create_api_tokens_table"
}

func (m *M20260924000003CreateAPITokensTable) Up() error {
	if facades.Schema().HasTable("api_tokens") {
		return nil
	}
	return facades.Schema().Create("api_tokens", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.String("name", 120)
		table.String("token_hash", 64)
		table.String("token_prefix", 24)
		table.Text("scopes_json")
		table.String("status", 24).Default("active")
		table.DateTimeTz("expires_at").Nullable()
		table.DateTimeTz("last_used_at").Nullable()
		table.String("last_used_ip", 64).Nullable()
		table.UnsignedBigInteger("usage_count").Default(0)
		table.DateTimeTz("revoked_at").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("token_hash")
		table.Index("user_id", "status", "id")
		table.Index("expires_at", "status")
	})
}

func (m *M20260924000003CreateAPITokensTable) Down() error {
	return facades.Schema().DropIfExists("api_tokens")
}
