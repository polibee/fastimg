package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/billing/providers"
)

var (
	ErrInvalidRefund     = errors.New("invalid refund")
	ErrRefundUnavailable = errors.New("refund unavailable")
)

func RefundableAmount(captured, refunded int64) int64 {
	if captured <= refunded {
		return 0
	}
	return captured - refunded
}

func ValidateRefundAmount(requested, refundable int64) error {
	if requested <= 0 || requested > refundable {
		return ErrInvalidRefund
	}
	return nil
}

type RefundService struct {
	gateways *providers.Registry
	ledger   *FinancialLedger
}

func NewRefundService(gateways *providers.Registry) *RefundService {
	return &RefundService{gateways: gateways, ledger: NewFinancialLedger()}
}

func (s *RefundService) Request(ctx context.Context, userID, orderID uint, amount int64, reason, idempotencyKey string) (*models.Refund, error) {
	if userID == 0 || orderID == 0 || strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrInvalidRefund
	}
	var order models.Order
	if err := facades.Orm().Query().Where("id = ? AND user_id = ?", orderID, userID).First(&order); err != nil {
		return nil, ErrOrderNotFound
	}
	var intent models.PaymentIntent
	if err := facades.Orm().Query().Where("order_id = ? AND status = ?", order.ID, "succeeded").First(&intent); err != nil {
		return nil, ErrRefundUnavailable
	}
	var refunds []models.Refund
	if err := facades.Orm().Query().Where("order_id = ? AND status IN (?, ?)", order.ID, "pending", "succeeded").Get(&refunds); err != nil {
		return nil, err
	}
	var refunded int64
	for _, item := range refunds {
		refunded += item.AmountMinor
	}
	if err := ValidateRefundAmount(amount, RefundableAmount(order.TotalAmountMinor, refunded)); err != nil {
		return nil, err
	}
	if existing, err := findRefundByIdempotency(idempotencyKey); err == nil && existing != nil {
		return existing, nil
	}
	refund := &models.Refund{OrderID: order.ID, PaymentTransactionID: 0, ProviderCode: intent.ProviderCode, AmountMinor: amount, Currency: order.Currency, Status: "pending", Reason: strings.TrimSpace(reason), IdempotencyKey: idempotencyKey, RequestedAt: time.Now().UTC()}
	if err := facades.Orm().Query().Create(refund); err != nil {
		return nil, err
	}
	gateway, err := s.gateways.Get(intent.ProviderCode)
	if err != nil {
		return nil, err
	}
	result, err := gateway.CreateRefund(ctx, providers.RefundRequest{OrderNo: order.PublicOrderNo, ProviderPaymentID: intent.ProviderPaymentID, AmountMinor: amount, Currency: order.Currency, IdempotencyKey: idempotencyKey, Reason: reason})
	if err != nil {
		_, _ = facades.Orm().Query().Where("id = ?", refund.ID).Update(map[string]any{"status": "failed"})
		return nil, fmt.Errorf("create refund: %w", err)
	}
	status := result.Status
	if status == "" {
		status = "pending"
	}
	if _, err := facades.Orm().Query().Where("id = ?", refund.ID).Update(map[string]any{"status": status, "provider_refund_id": result.ProviderRefundID}); err != nil {
		return nil, err
	}
	refund.Status, refund.ProviderRefundID = status, result.ProviderRefundID
	if status == "succeeded" {
		_ = s.ledger.Append(models.FinancialTransaction{OrderID: order.ID, ProviderCode: intent.ProviderCode, Type: "refund", Direction: "debit", AmountMinor: amount, Currency: order.Currency, SourceType: "refund", SourceID: fmt.Sprint(refund.ID), IdempotencyKey: "refund:" + idempotencyKey, OccurredAt: time.Now().UTC()})
	}
	return refund, nil
}

func findRefundByIdempotency(key string) (*models.Refund, error) {
	var refund models.Refund
	if err := facades.Orm().Query().Where("idempotency_key = ?", key).First(&refund); err != nil {
		return nil, err
	}
	return &refund, nil
}
