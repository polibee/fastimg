package migrations

import (
	"strings"
	"testing"
)

func TestEmailVerifiedAtMigrationBackfillsInOneDatabaseCommand(t *testing.T) {
	sql := emailVerifiedAtMigrationSQL()
	if strings.Count(sql, ";") != 1 {
		t.Fatalf("migration SQL must keep ALTER and backfill in one database command: %q", sql)
	}
	if !strings.Contains(sql, `ALTER TABLE "users" ADD COLUMN "email_verified_at" TIMESTAMP WITH TIME ZONE NULL`) {
		t.Fatalf("migration SQL does not add the nullable verification column: %q", sql)
	}
	if !strings.Contains(sql, `UPDATE "users" SET "email_verified_at" = "created_at" WHERE "email_verified_at" IS NULL`) {
		t.Fatalf("migration SQL does not preserve existing users as verified: %q", sql)
	}
}
