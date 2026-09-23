package migrations

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

type M20260923000001CreatePlansSubscriptionsUsageTables struct{}

func (m *M20260923000001CreatePlansSubscriptionsUsageTables) Signature() string {
	return "20260923000001_create_plans_subscriptions_usage_tables"
}

func (m *M20260923000001CreatePlansSubscriptionsUsageTables) Up() error {
	if !facades.Schema().HasTable("plans") {
		if err := facades.Schema().Create("plans", func(table schema.Blueprint) {
			table.ID()
			table.String("code", 64)
			table.String("name", 120)
			table.String("description", 255).Nullable()
			table.BigInteger("price_amount").Default(0)
			table.String("currency", 3).Default("CNY")
			table.String("billing_period", 32).Default("monthly")
			table.Text("entitlements_json")
			table.String("status", 32).Default("active")
			table.Integer("sort_order").Default(0)
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("code")
			table.Index("status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("subscriptions") {
		if err := facades.Schema().Create("subscriptions", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.UnsignedBigInteger("plan_id")
			table.String("status", 32).Default("active")
			table.DateTimeTz("starts_at").Nullable()
			table.DateTimeTz("ends_at").Nullable()
			table.DateTimeTz("canceled_at").Nullable()
			table.Text("entitlement_snapshot_json")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("user_id")
			table.Index("status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("usage_ledgers") {
		if err := facades.Schema().Create("usage_ledgers", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.String("resource_type", 32)
			table.BigInteger("delta")
			table.String("source_type", 32)
			table.String("source_id", 120)
			table.String("idempotency_key", 160)
			table.String("period_key", 32)
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("idempotency_key")
			table.Index("user_id", "resource_type", "period_key")
		}); err != nil {
			return err
		}
	}

	var plan models.Plan
	if err := facades.Orm().Query().Where("code = ?", "free").First(&plan); err != nil {
		encoded, encodeErr := json.Marshal(quota.DefaultFreeEntitlement())
		if encodeErr != nil {
			return encodeErr
		}
		if err := facades.Orm().Query().Create(&models.Plan{
			Code: "free", Name: "Free", Description: "Free media hosting baseline",
			PriceAmount: 0, Currency: "CNY", BillingPeriod: "monthly",
			EntitlementsJSON: string(encoded), Status: "active", SortOrder: 10,
		}); err != nil {
			return err
		}
	}
	for _, permission := range []models.Permission{
		{Name: "admin.plans.view", DisplayName: "Plans.view"},
		{Name: "admin.plans.create", DisplayName: "Plans.create"},
		{Name: "admin.plans.update", DisplayName: "Plans.update"},
		{Name: "admin.plans.delete", DisplayName: "Plans.delete"},
	} {
		exists, err := facades.Orm().Query().Model(&models.Permission{}).Where("name = ?", permission.Name).Exists()
		if err != nil {
			return err
		}
		if !exists {
			if err := facades.Orm().Query().Create(&permission); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *M20260923000001CreatePlansSubscriptionsUsageTables) Down() error {
	if err := facades.Schema().DropIfExists("usage_ledgers"); err != nil {
		return err
	}
	if err := facades.Schema().DropIfExists("subscriptions"); err != nil {
		return err
	}
	if err := facades.Schema().DropIfExists("plans"); err != nil {
		return err
	}
	for _, name := range []string{"admin.plans.view", "admin.plans.create", "admin.plans.update", "admin.plans.delete"} {
		if _, err := facades.Orm().Query().Table("permissions").Where("name = ?", name).Delete(); err != nil {
			return err
		}
	}
	return nil
}
