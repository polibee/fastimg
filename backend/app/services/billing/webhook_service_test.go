package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goravel/app/models"
	"goravel/app/services/billing/providers"
)

func TestPaymentTransactionIdentityRejectsCrossOrderReuse(t *testing.T) {
	event := providers.GatewayEvent{GatewayCode: "xcash", EventID: "event-1", ProviderPaymentTx: "tx-1", AmountMinor: 990, Currency: "USD"}
	existing := models.PaymentTransaction{OrderID: 8, PaymentIntentID: 11, ProviderCode: "xcash", ProviderTransaction: "tx-1", AmountMinor: 990, Currency: "USD"}
	require.True(t, paymentTransactionMatches(existing, event, 8, 11))
	require.False(t, paymentTransactionMatches(existing, event, 9, 11))
	require.False(t, paymentTransactionMatches(existing, event, 8, 12))
	mismatched := event
	mismatched.AmountMinor = 991
	require.False(t, paymentTransactionMatches(existing, mismatched, 8, 11))
}

func TestNormalizeGatewayPaymentStatus(t *testing.T) {
	cases := map[string]string{"succeeded": "succeeded", "completed": "succeeded", "confirmed": "succeeded", "pending": "pending", "confirming": "pending", "failed": "failed", "expired": "failed", "canceled": "failed"}
	for input, want := range cases {
		require.Equal(t, want, NormalizeGatewayPaymentStatus(input), input)
	}
}

func TestPaymentEventMustMatchLockedOrder(t *testing.T) {
	require.NoError(t, ValidatePaymentEvent(990, "USD", 990, "USD", "provider-payment-1"))
	require.Error(t, ValidatePaymentEvent(989, "USD", 990, "USD", "provider-payment-1"))
	require.Error(t, ValidatePaymentEvent(990, "EUR", 990, "USD", "provider-payment-1"))
	require.Error(t, ValidatePaymentEvent(990, "USD", 990, "USD", ""))
}
