package database

import (
	"testing"

	"goravel/database/migrations"
)

func TestBillingMigrationHasStableSignature(t *testing.T) {
	migration := migrations.M20260924000006CreateBillingTables{}
	if got, want := migration.Signature(), "20260924000006_create_billing_tables"; got != want {
		t.Fatalf("billing migration signature = %q, want %q", got, want)
	}
}
