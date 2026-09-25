package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260920000003AddRBACTimestamps struct{}

func (m *M20260920000003AddRBACTimestamps) Signature() string {
	return "20260920000003_add_rbac_timestamps"
}

func (m *M20260920000003AddRBACTimestamps) Up() error {
	for _, tableName := range []string{"roles", "permissions"} {
		columns := missingRBACTimestampColumns(tableName, facades.Schema().HasColumn)
		if len(columns) == 0 {
			continue
		}
		if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
			for _, column := range columns {
				table.DateTimeTz(column).Nullable()
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func missingRBACTimestampColumns(tableName string, hasColumn func(string, string) bool) []string {
	columns := make([]string, 0, 2)
	for _, column := range []string{"created_at", "updated_at"} {
		if !hasColumn(tableName, column) {
			columns = append(columns, column)
		}
	}
	return columns
}

func (m *M20260920000003AddRBACTimestamps) Down() error {
	return nil
}
