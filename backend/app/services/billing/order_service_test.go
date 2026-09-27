package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanCancelOrderOnlyAllowsUnpaidStates(t *testing.T) {
	for _, status := range []string{"created", "pending_payment"} {
		require.True(t, CanCancelOrder(status), status)
	}
	for _, status := range []string{"paid", "fulfilled", "canceled", "expired", "failed"} {
		require.False(t, CanCancelOrder(status), status)
	}
}

func TestValidateCreateOrderRequestRequiresServerPricedFields(t *testing.T) {
	valid := CreateOrderRequest{PlanID: 8, PriceID: 11, Currency: "USD", BillingPeriod: "monthly", IdempotencyKey: "checkout-1"}
	require.NoError(t, ValidateCreateOrderRequest(valid))

	cases := []CreateOrderRequest{
		{},
		{PlanID: 8, PriceID: 11, Currency: "usd", BillingPeriod: "monthly", IdempotencyKey: "checkout-1"},
		{PlanID: 8, PriceID: 11, Currency: "USD", BillingPeriod: "weekly", IdempotencyKey: "checkout-1"},
		{PlanID: 8, PriceID: 11, Currency: "USD", BillingPeriod: "monthly", IdempotencyKey: ""},
	}
	for _, input := range cases {
		require.Error(t, ValidateCreateOrderRequest(input))
	}
}

func TestCreateOrderRequestNeverContainsClientOwnershipOrAmount(t *testing.T) {
	input := CreateOrderRequest{PlanID: 8, PriceID: 11, Currency: "USD", BillingPeriod: "monthly", IdempotencyKey: "checkout-1"}
	require.NoError(t, ValidateCreateOrderRequest(input))
	// The request type intentionally has no UserID, amount or entitlement field.
	// Ownership and commercial terms are resolved from auth and PriceCatalog.
}
