package providers

import (
	"context"
	"errors"
	"fmt"
)

// RequestError is a safe summary of a provider HTTP failure. It deliberately
// carries only the status and provider error code/message, never credentials or
// request bodies.
type RequestError struct {
	GatewayCode     string
	StatusCode      int
	ProviderCode    string
	ProviderMessage string
	Retryable       bool
}

// NormalizeProviderError prevents an adapter-specific transport/protocol
// error from leaking as an opaque 500. Known safe RequestErrors pass through;
// unknown provider failures become a retryable, credential-safe summary.
func NormalizeProviderError(gatewayCode string, err error) error {
	if err == nil {
		return nil
	}
	var requestErr *RequestError
	if errors.As(err, &requestErr) {
		return err
	}
	return &RequestError{GatewayCode: gatewayCode, Retryable: true, ProviderMessage: "provider request failed"}
}

func (e *RequestError) Error() string {
	if e == nil {
		return "payment provider request failed"
	}
	if e.ProviderCode != "" {
		return fmt.Sprintf("%s provider request failed status=%d code=%s", e.GatewayCode, e.StatusCode, e.ProviderCode)
	}
	return fmt.Sprintf("%s provider request failed status=%d", e.GatewayCode, e.StatusCode)
}

type CreatePaymentRequest struct {
	OrderNo         string
	PaymentIntentID uint
	AmountMinor     int64
	Currency        string
	Description     string
	IdempotencyKey  string
}

type PaymentSession struct {
	GatewayCode       string
	ProviderPaymentID string
	ProviderOrderID   string
	Status            string
	CheckoutURL       string
	OccurredAt        int64
}

type QueryPaymentRequest struct {
	ProviderPaymentID string
}

type GatewayPayment struct {
	GatewayCode       string
	ProviderPaymentID string
	ProviderPaymentTx string
	Status            string
	AmountMinor       int64
	Currency          string
	OccurredAt        int64
}

type WebhookRequest struct {
	Headers map[string]string
	Body    []byte
}

type GatewayEvent struct {
	GatewayCode       string
	EventID           string
	EventType         string
	ProviderPaymentID string
	ProviderPaymentTx string
	Status            string
	AmountMinor       int64
	Currency          string
	OccurredAt        int64
}

type RefundRequest struct {
	OrderNo           string
	ProviderPaymentID string
	AmountMinor       int64
	Currency          string
	IdempotencyKey    string
	Reason            string
}

type GatewayRefund struct {
	GatewayCode      string
	ProviderRefundID string
	Status           string
	AmountMinor      int64
	Currency         string
	OccurredAt       int64
}

type PaymentGateway interface {
	CreatePayment(context.Context, CreatePaymentRequest) (PaymentSession, error)
	QueryPayment(context.Context, QueryPaymentRequest) (GatewayPayment, error)
	VerifyWebhook(context.Context, WebhookRequest) (GatewayEvent, error)
}

// RefundGateway is an optional provider capability. Refund support varies by
// channel and must not make every basic payment adapter implement placeholder
// methods just to satisfy the core gateway contract.
type RefundGateway interface {
	CreateRefund(context.Context, RefundRequest) (GatewayRefund, error)
	QueryRefund(context.Context, string) (GatewayRefund, error)
}
