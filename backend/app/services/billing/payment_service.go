package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/billing/providers"
)

var (
	ErrPaymentNotAllowed          = errors.New("payment is not allowed for this order")
	ErrPaymentIntentNotFound      = errors.New("payment intent not found")
	ErrPaymentIdempotencyConflict = errors.New("payment idempotency key belongs to another payment")
)

type StartPaymentRequest struct {
	GatewayCode string `json:"gateway_code"`
}

type PaymentService struct{ gateways *providers.Registry }

func NewPaymentService(gateways *providers.Registry) *PaymentService {
	return &PaymentService{gateways: gateways}
}

func (s *PaymentService) StartPayment(ctx context.Context, userID, orderID uint, input StartPaymentRequest, idempotencyKey string) (*models.PaymentIntent, error) {
	if userID == 0 || orderID == 0 || s.gateways == nil || input.GatewayCode == "" || idempotencyKey == "" {
		return nil, ErrInvalidOrderRequest
	}
	gateway, err := s.gateways.Get(input.GatewayCode)
	if err != nil {
		return nil, err
	}
	var order models.Order
	var intent models.PaymentIntent
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Where("id = ? AND user_id = ?", orderID, userID).First(&order); err != nil {
			return ErrOrderNotFound
		}
		if order.Status != "pending_payment" && order.Status != "created" {
			return ErrPaymentNotAllowed
		}
		if err := tx.Where("user_id = ? AND idempotency_key = ?", userID, idempotencyKey).First(&intent); err == nil {
			return ValidatePaymentIdempotency(intent, orderID, input.GatewayCode, idempotencyKey)
		}
		intent = models.PaymentIntent{OrderID: order.ID, UserID: userID, ProviderCode: input.GatewayCode, AmountMinor: order.TotalAmountMinor, Currency: order.Currency, Status: "created", AttemptNo: 1, IdempotencyKey: idempotencyKey}
		return tx.Create(&intent)
	})
	if err != nil {
		return nil, err
	}
	if intent.ProviderPaymentID != "" || intent.Status == "pending" {
		return &intent, nil
	}
	session, err := gateway.CreatePayment(ctx, providers.CreatePaymentRequest{OrderNo: order.PublicOrderNo, PaymentIntentID: intent.ID, AmountMinor: order.TotalAmountMinor, Currency: order.Currency, Description: "FastImg plan order", IdempotencyKey: idempotencyKey})
	if err != nil {
		_, _ = facades.Orm().Query().Model(&models.PaymentIntent{}).Where("id = ?", intent.ID).Update(map[string]any{"status": "failed", "failed_at": time.Now().UTC()})
		if errors.Is(err, providers.ErrGatewayUnavailable) {
			return nil, err
		}
		return nil, providers.NormalizeProviderError(input.GatewayCode, fmt.Errorf("create gateway payment: %w", err))
	}
	if _, err := facades.Orm().Query().Model(&models.PaymentIntent{}).Where("id = ?", intent.ID).Update(map[string]any{"status": session.Status, "provider_payment_id": session.ProviderPaymentID, "checkout_url": session.CheckoutURL}); err != nil {
		return nil, err
	}
	intent.Status, intent.ProviderPaymentID, intent.CheckoutURL = session.Status, session.ProviderPaymentID, session.CheckoutURL
	return &intent, nil
}

// ValidatePaymentIdempotency prevents a client from replaying a key created
// for a different order or provider and receiving the wrong checkout session.
func ValidatePaymentIdempotency(existing models.PaymentIntent, orderID uint, gatewayCode, key string) error {
	if existing.OrderID != orderID || existing.ProviderCode != gatewayCode || existing.IdempotencyKey != key {
		return ErrPaymentIdempotencyConflict
	}
	return nil
}
