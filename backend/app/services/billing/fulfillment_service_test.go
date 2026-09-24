package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFulfillmentRequiresVerifiedSucceededPayment(t *testing.T) {
	require.True(t, CanFulfillPayment("succeeded"))
	for _, status := range []string{"pending", "failed", "canceled", "requires_action"} {
		require.False(t, CanFulfillPayment(status), status)
	}
}

func TestFulfillmentStateKeepsPaidOrderWhenDeliveryFails(t *testing.T) {
	require.Equal(t, "fulfillment_pending", NextFulfillmentStatus("paid", false))
	require.Equal(t, "fulfilled", NextFulfillmentStatus("paid", true))
	require.Equal(t, "pending_payment", NextFulfillmentStatus("pending_payment", true))
}
