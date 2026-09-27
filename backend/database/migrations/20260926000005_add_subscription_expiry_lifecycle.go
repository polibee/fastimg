package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260926000005AddSubscriptionExpiryLifecycle struct{}

func (m *M20260926000005AddSubscriptionExpiryLifecycle) Signature() string {
	return "20260926000005_add_subscription_expiry_lifecycle"
}

func (m *M20260926000005AddSubscriptionExpiryLifecycle) Up() error {
	if facades.Schema().HasTable("subscriptions") && !facades.Schema().HasColumn("subscriptions", "grace_period_ends_at") {
		if err := facades.Schema().Table("subscriptions", func(table schema.Blueprint) {
			table.DateTimeTz("grace_period_ends_at").Nullable()
		}); err != nil {
			return err
		}
	}
	if facades.Schema().HasTable("subscription_notification_deliveries") {
		return nil
	}
	return facades.Schema().Create("subscription_notification_deliveries", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("subscription_id")
		table.UnsignedBigInteger("user_id")
		table.String("event_key", 160)
		table.String("kind", 32)
		table.String("channel", 24).Default("email")
		table.String("status", 24).Default("pending")
		table.Integer("attempts").Default(0)
		table.DateTimeTz("next_attempt_at").Nullable()
		table.DateTimeTz("sent_at").Nullable()
		table.Text("last_error").Nullable()
		table.DateTimeTz("created_at").Nullable()
		table.DateTimeTz("updated_at").Nullable()
		table.Unique("event_key")
		table.Index("subscription_id", "kind")
		table.Index("status", "next_attempt_at")
	})
}

func (m *M20260926000005AddSubscriptionExpiryLifecycle) Down() error {
	if err := facades.Schema().DropIfExists("subscription_notification_deliveries"); err != nil {
		return err
	}
	if facades.Schema().HasTable("subscriptions") && facades.Schema().HasColumn("subscriptions", "grace_period_ends_at") {
		return facades.Schema().Table("subscriptions", func(table schema.Blueprint) {
			table.DropColumn("grace_period_ends_at")
		})
	}
	return nil
}
