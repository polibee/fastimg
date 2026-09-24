package fake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/app/services/billing/providers"
)

func TestProviderCreatePaymentIsIdempotent(t *testing.T) {
	provider := New()
	request := providers.CreatePaymentRequest{OrderNo: "FST-1", PaymentIntentID: 42, AmountMinor: 990, Currency: "USD", IdempotencyKey: "pay-1"}

	first, err := provider.CreatePayment(context.Background(), request)
	require.NoError(t, err)
	second, err := provider.CreatePayment(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, first.ProviderPaymentID, second.ProviderPaymentID)
	require.Equal(t, first.CheckoutURL, second.CheckoutURL)
	require.Equal(t, "pending", first.Status)
}

func TestProviderTransitionCreatesVerifiedEvent(t *testing.T) {
	provider := New()
	payment, err := provider.CreatePayment(context.Background(), providers.CreatePaymentRequest{OrderNo: "FST-2", PaymentIntentID: 43, AmountMinor: 1200, Currency: "USD", IdempotencyKey: "pay-2"})
	require.NoError(t, err)

	event, err := provider.Transition(payment.ProviderPaymentID, "succeeded")
	require.NoError(t, err)
	require.Equal(t, "fake.payment.succeeded", event.EventType)
	require.Equal(t, "fake-event-"+payment.ProviderPaymentID, event.EventID)
}
