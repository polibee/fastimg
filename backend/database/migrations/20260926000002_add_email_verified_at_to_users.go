package migrations

import "goravel/app/facades"

type M20260926000002AddEmailVerifiedAtToUsers struct{}

func (m *M20260926000002AddEmailVerifiedAtToUsers) Signature() string {
	return "20260926000002_add_email_verified_at_to_users"
}

func (m *M20260926000002AddEmailVerifiedAtToUsers) Up() error {
	if !facades.Schema().HasTable("users") || facades.Schema().HasColumn("users", "email_verified_at") {
		return nil
	}
	_, err := facades.Orm().Query().Exec(emailVerifiedAtMigrationSQL())
	return err
}

func (m *M20260926000002AddEmailVerifiedAtToUsers) Down() error { return nil }

// Keep the schema change and the compatibility backfill on one PostgreSQL
// connection. Goravel's schema builder can leave an ALTER TABLE transaction
// open while the following ORM update is acquired from another pool
// connection, which self-blocks on PostgreSQL's relation lock.
func emailVerifiedAtMigrationSQL() string {
	return `ALTER TABLE "users" ADD COLUMN "email_verified_at" TIMESTAMP WITH TIME ZONE NULL; UPDATE "users" SET "email_verified_at" = "created_at" WHERE "email_verified_at" IS NULL`
}
