package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goravel/app/models"
)

func TestPaymentIdempotencyMustStayWithinOrderAndGateway(t *testing.T) {
	existing := models.PaymentIntent{OrderID: 8, ProviderCode: "xcash", IdempotencyKey: "checkout-1"}
	require.NoError(t, ValidatePaymentIdempotency(existing, 8, "xcash", "checkout-1"))
	require.ErrorIs(t, ValidatePaymentIdempotency(existing, 9, "xcash", "checkout-1"), ErrPaymentIdempotencyConflict)
	require.ErrorIs(t, ValidatePaymentIdempotency(existing, 8, "paypal", "checkout-1"), ErrPaymentIdempotencyConflict)
	require.ErrorIs(t, ValidatePaymentIdempotency(existing, 8, "xcash", "checkout-2"), ErrPaymentIdempotencyConflict)
}
