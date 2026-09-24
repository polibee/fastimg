package fake

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"goravel/app/services/billing/providers"
)

var ErrPaymentNotFound = errors.New("fake payment not found")
var ErrUnsupportedTransition = errors.New("fake payment transition is unsupported")

type payment struct {
	request providers.CreatePaymentRequest
	status  string
}

type Provider struct {
	mu    sync.Mutex
	byKey map[string]payment
	byID  map[string]string
}

func New() *Provider {
	return &Provider{byKey: map[string]payment{}, byID: map[string]string{}}
}

func (p *Provider) CreatePayment(_ context.Context, request providers.CreatePaymentRequest) (providers.PaymentSession, error) {
	if request.IdempotencyKey == "" || request.PaymentIntentID == 0 {
		return providers.PaymentSession{}, errors.New("fake payment requires idempotency key and payment intent")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if existingID, ok := p.byID[request.IdempotencyKey]; ok {
		existing := p.byKey[request.IdempotencyKey]
		return p.session(existingID, existing.status), nil
	}
	providerID := fmt.Sprintf("fake-payment-%d", request.PaymentIntentID)
	p.byID[request.IdempotencyKey] = providerID
	p.byKey[request.IdempotencyKey] = payment{request: request, status: "pending"}
	return p.session(providerID, "pending"), nil
}

func (p *Provider) QueryPayment(_ context.Context, request providers.QueryPaymentRequest) (providers.GatewayPayment, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for key, providerID := range p.byID {
		if providerID != request.ProviderPaymentID {
			continue
		}
		item := p.byKey[key]
		return providers.GatewayPayment{GatewayCode: "fake", ProviderPaymentID: providerID, Status: item.status, AmountMinor: item.request.AmountMinor, Currency: item.request.Currency, OccurredAt: time.Now().Unix()}, nil
	}
	return providers.GatewayPayment{}, ErrPaymentNotFound
}

func (p *Provider) VerifyWebhook(_ context.Context, request providers.WebhookRequest) (providers.GatewayEvent, error) {
	return providers.GatewayEvent{}, errors.New("fake webhook verification requires Transition event transport")
}

func (p *Provider) CreateRefund(context.Context, providers.RefundRequest) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, errors.New("fake refunds are not implemented yet")
}

func (p *Provider) QueryRefund(context.Context, string) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, errors.New("fake refunds are not implemented yet")
}

func (p *Provider) Transition(providerPaymentID, status string) (providers.GatewayEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for key, id := range p.byID {
		if id != providerPaymentID {
			continue
		}
		if status != "pending" && status != "succeeded" && status != "failed" && status != "canceled" {
			return providers.GatewayEvent{}, ErrUnsupportedTransition
		}
		item := p.byKey[key]
		item.status = status
		p.byKey[key] = item
		return providers.GatewayEvent{
			GatewayCode: "fake", EventID: "fake-event-" + providerPaymentID,
			EventType: "fake.payment." + status, ProviderPaymentID: providerPaymentID,
			Status: status, AmountMinor: item.request.AmountMinor, Currency: item.request.Currency,
			OccurredAt: time.Now().Unix(),
		}, nil
	}
	return providers.GatewayEvent{}, ErrPaymentNotFound
}

func (p *Provider) session(providerID, status string) providers.PaymentSession {
	return providers.PaymentSession{GatewayCode: "fake", ProviderPaymentID: providerID, Status: status, CheckoutURL: "/checkout/fake/" + providerID, OccurredAt: time.Now().Unix()}
}
