package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260926000003CreateEmailVerificationTokensTable struct{}

func (m *M20260926000003CreateEmailVerificationTokensTable) Signature() string {
	return "20260926000003_create_email_verification_tokens_table"
}

func (m *M20260926000003CreateEmailVerificationTokensTable) Up() error {
	if facades.Schema().HasTable("email_verification_tokens") {
		return nil
	}
	return facades.Schema().Create("email_verification_tokens", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.String("token_hash", 64)
		table.DateTimeTz("expires_at")
		table.DateTimeTz("consumed_at").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("token_hash")
		table.Index("user_id", "expires_at")
	})
}

func (m *M20260926000003CreateEmailVerificationTokensTable) Down() error {
	return facades.Schema().DropIfExists("email_verification_tokens")
}
