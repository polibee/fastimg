package migrations

import "testing"

func TestBackupJobsMigrationHasStableSignature(t *testing.T) {

	migration := &M20260926000004CreateBackupJobsTable{}
	if got := migration.Signature(); got != "20260926000004_create_backup_jobs_table" {
		t.Fatalf("signature = %q", got)
	}
}

