package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
