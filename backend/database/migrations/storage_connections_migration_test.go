package migrations

import "testing"

func TestStorageConnectionsMigrationHasStableSignature(t *testing.T) {
	migration := &M20260927000001CreateStorageConnectionsTable{}
	if got := migration.Signature(); got != "20260927000001_create_storage_connections_table" {
		t.Fatalf("signature = %q", got)
	}
}
