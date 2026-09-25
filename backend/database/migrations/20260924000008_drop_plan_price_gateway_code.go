package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260924000008DropPlanPriceGatewayCode removes the obsolete price-to-gateway
// coupling. A price is gateway-neutral; the member selects an enabled provider
// and that choice is recorded on the PaymentIntent.
type M20260924000008DropPlanPriceGatewayCode struct{}

func (m *M20260924000008DropPlanPriceGatewayCode) Signature() string {
	return "20260924000008_drop_plan_price_gateway_code"
}

func (m *M20260924000008DropPlanPriceGatewayCode) Up() error {
	if !facades.Schema().HasTable("plan_prices") || !facades.Schema().HasColumn("plan_prices", "gateway_code") {
		return nil
	}
	return facades.Schema().Table("plan_prices", func(table schema.Blueprint) {
		table.DropColumn("gateway_code")
	})
}

func (m *M20260924000008DropPlanPriceGatewayCode) Down() error {
	// Do not restore the obsolete coupling during rollback.
	return nil
}
