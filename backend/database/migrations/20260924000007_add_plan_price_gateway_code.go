package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260924000007AddPlanPriceGatewayCode binds each immutable price version to
// the gateway selected by an administrator. Secrets remain environment-only.
type M20260924000007AddPlanPriceGatewayCode struct{}

func (m *M20260924000007AddPlanPriceGatewayCode) Signature() string {
	return "20260924000007_add_plan_price_gateway_code"
}

func (m *M20260924000007AddPlanPriceGatewayCode) Up() error {
	if !facades.Schema().HasTable("plan_prices") || facades.Schema().HasColumn("plan_prices", "gateway_code") {
		return nil
	}
	return facades.Schema().Table("plan_prices", func(table schema.Blueprint) {
		table.String("gateway_code", 32).Default("fake")
	})
}

func (m *M20260924000007AddPlanPriceGatewayCode) Down() error {
	if !facades.Schema().HasTable("plan_prices") || !facades.Schema().HasColumn("plan_prices", "gateway_code") {
		return nil
	}
	return facades.Schema().Table("plan_prices", func(table schema.Blueprint) {
		table.DropColumn("gateway_code")
	})
}
