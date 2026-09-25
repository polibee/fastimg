package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260921000001ReplaceUserActiveWithStatus struct{}

func (m *M20260921000001ReplaceUserActiveWithStatus) Signature() string {
	return "20260921000001_replace_user_active_with_status"
}

func (m *M20260921000001ReplaceUserActiveWithStatus) Up() error {
	addStatus, migrateLegacyActive := userStatusMigrationActions(facades.Schema().HasColumn)
	if addStatus {
		if err := facades.Schema().Table("users", func(table schema.Blueprint) {
			table.String("status").Default("active")
		}); err != nil {
			return err
		}
	}
	if !migrateLegacyActive {
		return nil
	}
	if _, err := facades.Orm().Query().Table("users").Where("is_active = ?", false).Update("status", "disabled"); err != nil {
		return err
	}
	return facades.Schema().Table("users", func(table schema.Blueprint) {
		table.DropColumn("is_active")
	})
}

func userStatusMigrationActions(hasColumn func(string, string) bool) (addStatus, migrateLegacyActive bool) {
	return !hasColumn("users", "status"), hasColumn("users", "is_active")
}

func (m *M20260921000001ReplaceUserActiveWithStatus) Down() error {
	// The base users migration also creates status, so this migration cannot
	// safely infer whether dropping it would remove pre-existing application data.
	return nil
}
