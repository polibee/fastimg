package providers

import "context"

type CreatePaymentRequest struct {
	OrderNo         string
	PaymentIntentID uint
	AmountMinor     int64
	Currency        string
	Description     string
	CallbackURL     string
	ReturnURL       string
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
	CreateRefund(context.Context, RefundRequest) (GatewayRefund, error)
	QueryRefund(context.Context, string) (GatewayRefund, error)
}
