package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260924000006CreateBillingTables creates the payment-domain facts. JSON
// snapshots use Text because the current Goravel schema abstraction does not
// expose a portable JSONB builder; the service validates every snapshot before
// persistence and PostgreSQL deployments can add JSONB-specific indexes later.
type M20260924000006CreateBillingTables struct{}

func (m *M20260924000006CreateBillingTables) Signature() string {
	return "20260924000006_create_billing_tables"
}

func (m *M20260924000006CreateBillingTables) Up() error {
	if !facades.Schema().HasTable("plan_prices") {
		if err := facades.Schema().Create("plan_prices", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("plan_id")
			table.String("version", 40)
			table.String("currency", 3)
			table.BigInteger("amount_minor")
			 table.String("billing_period", 16)
			table.Integer("trial_days").Default(0)
			table.String("status", 16).Default("draft")
			table.DateTimeTz("effective_from")
			table.DateTimeTz("effective_to").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("plan_id", "version")
			table.Index("plan_id", "currency", "billing_period", "status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("orders") {
		if err := facades.Schema().Create("orders", func(table schema.Blueprint) {
			table.ID()
			table.String("public_order_no", 40)
			table.UnsignedBigInteger("user_id")
			table.String("status", 32)
			table.String("currency", 3)
			table.BigInteger("subtotal_amount_minor")
			table.BigInteger("discount_amount_minor").Default(0)
			table.BigInteger("tax_amount_minor").Default(0)
			table.BigInteger("total_amount_minor")
			table.Text("price_snapshot")
			table.Text("customer_snapshot")
			table.String("idempotency_key", 120)
			table.DateTimeTz("expires_at").Nullable()
			table.DateTimeTz("paid_at").Nullable()
			table.DateTimeTz("fulfilled_at").Nullable()
			table.DateTimeTz("canceled_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("public_order_no")
			table.Unique("user_id", "idempotency_key")
			table.Index("user_id", "status", "created_at")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("order_items") {
		if err := facades.Schema().Create("order_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.String("product_type", 32)
			table.UnsignedBigInteger("product_id")
			table.String("product_version", 80)
			table.BigInteger("quantity")
			table.BigInteger("unit_amount_minor")
			table.BigInteger("discount_amount_minor").Default(0)
			table.BigInteger("total_amount_minor")
			table.Text("entitlement_snapshot")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("order_id")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("payment_intents") {
		if err := facades.Schema().Create("payment_intents", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.UnsignedBigInteger("user_id")
			table.String("provider_code", 32)
			table.BigInteger("amount_minor")
			table.String("currency", 3)
			table.String("status", 32)
			table.Integer("attempt_no").Default(0)
			table.String("provider_payment_id", 160).Nullable()
			table.Text("checkout_url").Nullable()
			table.String("idempotency_key", 120)
			table.DateTimeTz("expires_at").Nullable()
			table.DateTimeTz("succeeded_at").Nullable()
			table.DateTimeTz("failed_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("user_id", "idempotency_key")
			table.Index("order_id", "status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("payment_attempts") {
		if err := facades.Schema().Create("payment_attempts", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("payment_intent_id")
			table.String("provider_code", 32)
			table.String("idempotency_key", 120)
			table.String("provider_payment_id", 160).Nullable()
			table.String("status", 32)
			table.String("request_hash", 64).Nullable()
			table.String("response_hash", 64).Nullable()
			table.DateTimeTz("occurred_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("provider_code", "idempotency_key")
			table.Index("payment_intent_id", "status")
		}); err != nil {
			return err
		}
	}
	return createBillingTailTables()
}

func createBillingTailTables() error {
	if !facades.Schema().HasTable("payment_transactions") {
		if err := facades.Schema().Create("payment_transactions", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.UnsignedBigInteger("payment_intent_id")
			table.String("provider_code", 32)
			table.String("type", 24)
			table.String("direction", 12)
			table.BigInteger("amount_minor")
			table.String("currency", 3)
			table.String("status", 24)
			table.String("provider_transaction_id", 160).Nullable()
			table.String("provider_event_id", 160).Nullable()
			table.DateTimeTz("occurred_at")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("provider_code", "provider_transaction_id")
			table.Index("order_id", "occurred_at")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("payment_webhook_events") {
		if err := facades.Schema().Create("payment_webhook_events", func(table schema.Blueprint) {
			table.ID()
			table.String("provider_code", 32)
			table.String("event_id", 160)
			table.String("event_type", 120)
			table.Boolean("signature_valid").Default(false)
			table.String("payload_hash", 64)
			table.String("payload_reference", 255).Nullable()
			table.String("processing_status", 24)
			table.Integer("retry_count").Default(0)
			table.DateTimeTz("processed_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("provider_code", "event_id")
			table.Index("provider_code", "processing_status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("refunds") {
		if err := facades.Schema().Create("refunds", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.UnsignedBigInteger("payment_transaction_id")
			table.String("provider_code", 32)
			table.BigInteger("amount_minor")
			table.String("currency", 3)
			table.String("status", 24)
			table.String("reason", 255).Nullable()
			table.String("provider_refund_id", 160).Nullable()
			table.String("idempotency_key", 160)
			table.DateTimeTz("requested_at")
			table.DateTimeTz("completed_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("order_id", "status")
			table.Unique("idempotency_key")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("financial_transactions") {
		if err := facades.Schema().Create("financial_transactions", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id").Nullable()
			table.String("provider_code", 32).Nullable()
			table.String("type", 24)
			table.String("direction", 12)
			table.BigInteger("amount_minor")
			table.String("currency", 3)
			table.String("source_type", 32)
			table.String("source_id", 160)
			table.String("idempotency_key", 160)
			table.DateTimeTz("occurred_at")
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("idempotency_key")
			table.Index("type", "occurred_at")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("fulfillment_tasks") {
		return facades.Schema().Create("fulfillment_tasks", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.String("status", 24)
			table.Integer("attempts").Default(0)
			table.Text("last_error").Nullable()
			table.DateTimeTz("available_at").Nullable()
			table.DateTimeTz("completed_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("order_id")
			table.Index("status", "available_at")
		})
	}
	return nil
}

func (m *M20260924000006CreateBillingTables) Down() error {
	for _, table := range []string{"fulfillment_tasks", "financial_transactions", "refunds", "payment_webhook_events", "payment_transactions", "payment_attempts", "payment_intents", "order_items", "orders", "plan_prices"} {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}
	return nil
}
