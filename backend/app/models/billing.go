package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

// PlanPrice is an immutable commercial price version. A new amount or period
// must create a new row instead of changing a price already used by an order.
type PlanPrice struct {
	orm.Model
	PlanID        uint       `json:"plan_id"`
	Version       string     `json:"version"`
	Currency      string     `json:"currency"`
	AmountMinor   int64      `json:"amount_minor"`
	BillingPeriod string     `json:"billing_period"`
	TrialDays     int        `json:"trial_days"`
	Status        string     `json:"status"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to"`
}

type Order struct {
	orm.Model
	PublicOrderNo       string     `json:"public_order_no"`
	UserID              uint       `json:"user_id"`
	Status              string     `json:"status"`
	Currency            string     `json:"currency"`
	SubtotalAmountMinor int64      `json:"subtotal_amount_minor"`
	DiscountAmountMinor int64      `json:"discount_amount_minor"`
	TaxAmountMinor      int64      `json:"tax_amount_minor"`
	TotalAmountMinor    int64      `json:"total_amount_minor"`
	PriceSnapshot       string     `json:"-"`
	CustomerSnapshot    string     `json:"-"`
	IdempotencyKey      string     `json:"-"`
	ExpiresAt           *time.Time `json:"expires_at"`
	PaidAt              *time.Time `json:"paid_at"`
	FulfilledAt         *time.Time `json:"fulfilled_at"`
	CanceledAt          *time.Time `json:"canceled_at"`
}

type OrderItem struct {
	orm.Model
	OrderID             uint   `json:"order_id"`
	ProductType         string `json:"product_type"`
	ProductID           uint   `json:"product_id"`
	ProductVersion      string `json:"product_version"`
	Quantity            int64  `json:"quantity"`
	UnitAmountMinor     int64  `json:"unit_amount_minor"`
	DiscountAmountMinor int64  `json:"discount_amount_minor"`
	TotalAmountMinor    int64  `json:"total_amount_minor"`
	EntitlementSnapshot string `json:"-"`
}

type PaymentIntent struct {
	orm.Model
	OrderID           uint       `json:"order_id"`
	UserID            uint       `json:"user_id"`
	ProviderCode      string     `json:"provider_code"`
	AmountMinor       int64      `json:"amount_minor"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	AttemptNo         int        `json:"attempt_no"`
	ProviderPaymentID string     `json:"provider_payment_id"`
	CheckoutURL       string     `json:"checkout_url"`
	IdempotencyKey    string     `json:"-"`
	ExpiresAt         *time.Time `json:"expires_at"`
	SucceededAt       *time.Time `json:"succeeded_at"`
	FailedAt          *time.Time `json:"failed_at"`
}

type PaymentAttempt struct {
	orm.Model
	PaymentIntentID   uint       `json:"payment_intent_id"`
	ProviderCode      string     `json:"provider_code"`
	IdempotencyKey    string     `json:"-"`
	ProviderPaymentID string     `json:"provider_payment_id"`
	Status            string     `json:"status"`
	RequestHash       string     `json:"request_hash"`
	ResponseHash      string     `json:"response_hash"`
	OccurredAt        *time.Time `json:"occurred_at"`
}

type PaymentTransaction struct {
	orm.Model
	OrderID             uint      `json:"order_id"`
	PaymentIntentID     uint      `json:"payment_intent_id"`
	ProviderCode        string    `json:"provider_code"`
	Type                string    `json:"type"`
	Direction           string    `json:"direction"`
	AmountMinor         int64     `json:"amount_minor"`
	Currency            string    `json:"currency"`
	Status              string    `json:"status"`
	ProviderTransaction string    `json:"provider_transaction_id"`
	ProviderEventID     string    `json:"provider_event_id"`
	OccurredAt          time.Time `json:"occurred_at"`
}

type PaymentWebhookEvent struct {
	orm.Model
	ProviderCode     string     `json:"provider_code"`
	EventID          string     `json:"event_id"`
	EventType        string     `json:"event_type"`
	SignatureValid   bool       `json:"signature_valid"`
	PayloadHash      string     `json:"payload_hash"`
	PayloadReference string     `json:"payload_reference"`
	ProcessingStatus string     `json:"processing_status"`
	RetryCount       int        `json:"retry_count"`
	ProcessedAt      *time.Time `json:"processed_at"`
}

type Refund struct {
	orm.Model
	OrderID              uint       `json:"order_id"`
	PaymentTransactionID uint       `json:"payment_transaction_id"`
	ProviderCode         string     `json:"provider_code"`
	AmountMinor          int64      `json:"amount_minor"`
	Currency             string     `json:"currency"`
	Status               string     `json:"status"`
	Reason               string     `json:"reason"`
	ProviderRefundID     string     `json:"provider_refund_id"`
	RequestedAt          time.Time  `json:"requested_at"`
	CompletedAt          *time.Time `json:"completed_at"`
}

type FinancialTransaction struct {
	orm.Model
	OrderID        uint      `json:"order_id"`
	ProviderCode   string    `json:"provider_code"`
	Type           string    `json:"type"`
	Direction      string    `json:"direction"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	SourceType     string    `json:"source_type"`
	SourceID       string    `json:"source_id"`
	IdempotencyKey string    `json:"-"`
	OccurredAt     time.Time `json:"occurred_at"`
}

type FulfillmentTask struct {
	orm.Model
	OrderID     uint       `json:"order_id"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"last_error"`
	AvailableAt *time.Time `json:"available_at"`
	CompletedAt *time.Time `json:"completed_at"`
}
