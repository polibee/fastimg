package migrations

import (
	"goravel/app/facades"
	"goravel/app/models"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260927000001CreateStorageConnectionsTable struct{}

func (m *M20260927000001CreateStorageConnectionsTable) Signature() string {
	return "20260927000001_create_storage_connections_table"
}

func (m *M20260927000001CreateStorageConnectionsTable) Up() error {
	if !facades.Schema().HasTable("storage_connections") {
		if err := facades.Schema().Create("storage_connections", func(table schema.Blueprint) {
			table.ID()
			table.String("provider_code", 32)
			table.String("name", 120)
			table.Boolean("enabled").Default(false)
			table.Boolean("is_primary").Default(false)
			table.String("status", 24).Default(models.StorageConnectionDisabled)
			table.Text("config_encrypted").Nullable()
			table.String("public_base_url", 512).Nullable()
			table.String("path_prefix", 255).Nullable()
			table.String("default_visibility", 16).Default("public")
			table.Integer("signed_url_ttl_seconds").Default(3600)
			table.DateTimeTz("last_checked_at").Nullable()
			table.DateTimeTz("last_success_at").Nullable()
			table.String("last_error_code", 80).Nullable()
			table.Text("last_error_message").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("provider_code", "name")
			table.Index("enabled", "is_primary")
		}); err != nil {
			return err
		}
	}

	localID, err := ensureLocalStorageConnection()
	if err != nil {
		return err
	}

	if facades.Schema().HasTable("storage_objects") && !facades.Schema().HasColumn("storage_objects", "storage_connection_id") {
		if err := facades.Schema().Table("storage_objects", func(table schema.Blueprint) {
			table.UnsignedBigInteger("storage_connection_id").Nullable()
		}); err != nil {
			return err
		}
	}
	if facades.Schema().HasTable("storage_objects") && facades.Schema().HasColumn("storage_objects", "storage_connection_id") {
		if _, err := facades.Orm().Query().Table("storage_objects").Where("storage_connection_id IS NULL").Update("storage_connection_id", localID); err != nil {
			return err
		}
		if facades.Schema().HasColumn("storage_objects", "provider") {
			if err := facades.Schema().Table("storage_objects", func(table schema.Blueprint) {
				table.DropColumn("provider")
			}); err != nil {
				return err
			}
		}
	}

	if !facades.Schema().HasTable("storage_usage_snapshots") {
		if err := facades.Schema().Create("storage_usage_snapshots", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("storage_connection_id")
			table.String("source", 32)
			table.BigInteger("object_count").Default(0)
			table.BigInteger("total_bytes").Default(0)
			table.BigInteger("egress_bytes").Nullable()
			table.BigInteger("request_count").Nullable()
			table.DateTimeTz("captured_at")
			table.String("error_code", 80).Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("storage_connection_id", "captured_at")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("storage_health_checks") {
		return facades.Schema().Create("storage_health_checks", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("storage_connection_id")
			table.String("check_type", 32)
			table.String("status", 24)
			table.Integer("latency_ms").Nullable()
			table.String("error_code", 80).Nullable()
			table.Text("error_message").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("storage_connection_id", "created_at")
		})
	}
	return nil
}

func ensureLocalStorageConnection() (uint, error) {
	var connections []models.StorageConnection
	if err := facades.Orm().Query().Where("provider_code = ? AND name = ?", models.StorageProviderLocal, "Local").Get(&connections); err != nil {
		return 0, err
	}
	if len(connections) > 0 {
		return connections[0].ID, nil
	}
	connection := models.StorageConnection{
		ProviderCode:        models.StorageProviderLocal,
		Name:                "Local",
		Enabled:             true,
		IsPrimary:           true,
		Status:              models.StorageConnectionHealthy,
		DefaultVisibility:   "public",
		SignedURLTTLSeconds: 3600,
	}
	if err := facades.Orm().Query().Create(&connection); err != nil {
		return 0, err
	}
	return connection.ID, nil
}

func (m *M20260927000001CreateStorageConnectionsTable) Down() error {
	if facades.Schema().HasTable("storage_health_checks") {
		if err := facades.Schema().DropIfExists("storage_health_checks"); err != nil {
			return err
		}
	}
	if facades.Schema().HasTable("storage_usage_snapshots") {
		if err := facades.Schema().DropIfExists("storage_usage_snapshots"); err != nil {
			return err
		}
	}
	if facades.Schema().HasTable("storage_objects") && facades.Schema().HasColumn("storage_objects", "storage_connection_id") {
		if err := facades.Schema().Table("storage_objects", func(table schema.Blueprint) {
			table.DropColumn("storage_connection_id")
		}); err != nil {
			return err
		}
	}
	return facades.Schema().DropIfExists("storage_connections")
}
