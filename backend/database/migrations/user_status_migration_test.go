package migrations

import "testing"

func TestUserStatusMigrationSkipsAlreadyCurrentSchema(t *testing.T) {
	hasColumn := func(table, column string) bool {
		return table == "users" && column == "status"
	}
	addStatus, migrateLegacyActive := userStatusMigrationActions(hasColumn)
	if addStatus || migrateLegacyActive {
		t.Fatalf("current users schema should need no changes; addStatus=%v migrateLegacyActive=%v", addStatus, migrateLegacyActive)
	}
}

func TestUserStatusMigrationConvertsLegacyActiveSchema(t *testing.T) {
	hasColumn := func(table, column string) bool {
		return table == "users" && column == "is_active"
	}
	addStatus, migrateLegacyActive := userStatusMigrationActions(hasColumn)
	if !addStatus || !migrateLegacyActive {
		t.Fatalf("legacy users schema should add status and migrate is_active; addStatus=%v migrateLegacyActive=%v", addStatus, migrateLegacyActive)
	}
}
