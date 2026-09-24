package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/billing/providers"
)

var (
	ErrInvalidPaymentEvent  = errors.New("invalid payment event")
	ErrPaymentEventRejected = errors.New("payment event rejected")
)

func NormalizeGatewayPaymentStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "succeeded", "completed", "confirmed", "finished", "settled":
		return "succeeded"
	case "pending", "waiting", "confirming", "detected", "requires_action":
		return "pending"
	case "failed", "expired", "canceled", "cancelled", "denied", "reversed":
		return "failed"
	default:
		return "pending"
	}
}

func ValidatePaymentEvent(eventAmount int64, eventCurrency string, orderAmount int64, orderCurrency, providerPaymentID string) error {
	if providerPaymentID == "" || eventAmount < 0 || eventAmount != orderAmount || !strings.EqualFold(eventCurrency, orderCurrency) {
		return ErrInvalidPaymentEvent
	}
	return nil
}

type WebhookService struct{ gateways *providers.Registry }

func NewWebhookService(gateways *providers.Registry) *WebhookService {
	return &WebhookService{gateways: gateways}
}

// Ingest verifies a provider event, persists it once and applies only the
// normalized payment fact. Heavy fulfillment runs after this transaction.
func (s *WebhookService) Ingest(ctx context.Context, gatewayCode string, request providers.WebhookRequest) (providers.GatewayEvent, bool, error) {
	if s.gateways == nil {
		return providers.GatewayEvent{}, false, providers.ErrGatewayUnavailable
	}
	gateway, err := s.gateways.Get(gatewayCode)
	if err != nil {
		return providers.GatewayEvent{}, false, err
	}
	event, err := gateway.VerifyWebhook(ctx, request)
	if err != nil {
		return providers.GatewayEvent{}, false, err
	}
	if event.GatewayCode == "" {
		event.GatewayCode = gatewayCode
	}
	if event.EventID == "" || !strings.EqualFold(event.GatewayCode, gatewayCode) {
		return providers.GatewayEvent{}, false, ErrInvalidPaymentEvent
	}
	hash := sha256.Sum256(request.Body)
	payloadHash := hex.EncodeToString(hash[:])
	var duplicate bool
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		var stored models.PaymentWebhookEvent
		if exists, findErr := tx.Model(&models.PaymentWebhookEvent{}).Where("provider_code = ? AND event_id = ?", gatewayCode, event.EventID).Exists(); findErr != nil {
			return findErr
		} else if exists {
			duplicate = true
			return nil
		}
		stored = models.PaymentWebhookEvent{ProviderCode: gatewayCode, EventID: event.EventID, EventType: event.EventType, SignatureValid: true, PayloadHash: payloadHash, ProcessingStatus: "received"}
		if err := tx.Create(&stored); err != nil {
			return err
		}
		return applyGatewayEvent(tx, event)
	})
	if err != nil {
		return event, duplicate, err
	}
	if !duplicate && NormalizeGatewayPaymentStatus(event.Status) == "succeeded" {
		var intent models.PaymentIntent
		if err := facades.Orm().Query().Where("provider_code = ? AND provider_payment_id = ?", gatewayCode, event.ProviderPaymentID).First(&intent); err == nil {
			_ = NewFulfillmentService().EnqueuePaidOrder(intent.OrderID)
		}
	}
	return event, duplicate, nil
}

func applyGatewayEvent(tx orm.Query, event providers.GatewayEvent) error {
	var intent models.PaymentIntent
	if err := tx.Where("provider_code = ? AND provider_payment_id = ?", event.GatewayCode, event.ProviderPaymentID).First(&intent); err != nil {
		return ErrPaymentEventRejected
	}
	var order models.Order
	if err := tx.Where("id = ?", intent.OrderID).First(&order); err != nil {
		return ErrOrderNotFound
	}
	if err := ValidatePaymentEvent(event.AmountMinor, event.Currency, order.TotalAmountMinor, order.Currency, event.ProviderPaymentID); err != nil {
		return err
	}
	normalized := NormalizeGatewayPaymentStatus(event.Status)
	now := time.Now().UTC()
	if normalized == "succeeded" {
		transactionID := event.ProviderPaymentTx
		if transactionID == "" {
			transactionID = event.EventID
		}
		transaction := models.PaymentTransaction{OrderID: order.ID, PaymentIntentID: intent.ID, ProviderCode: event.GatewayCode, Type: "payment", Direction: "credit", AmountMinor: event.AmountMinor, Currency: event.Currency, Status: "settled", ProviderTransaction: transactionID, ProviderEventID: event.EventID, OccurredAt: now}
		if err := tx.Create(&transaction); err != nil {
			return fmt.Errorf("append payment transaction: %w", err)
		}
		if _, err := tx.Where("id = ?", intent.ID).Update(map[string]any{"status": "succeeded", "succeeded_at": now}); err != nil {
			return err
		}
		if _, err := tx.Where("id = ?", order.ID).Update(map[string]any{"status": "paid", "paid_at": now}); err != nil {
			return err
		}
		return nil
	}
	if _, err := tx.Where("id = ?", intent.ID).Update(map[string]any{"status": normalized}); err != nil {
		return err
	}
	return nil
}
