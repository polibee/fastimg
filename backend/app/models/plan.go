package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type Plan struct {
	orm.Model
	Code             string `json:"code"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	PriceAmount      int64  `json:"price_amount"`
	Currency         string `json:"currency"`
	BillingPeriod    string `json:"billing_period"`
	EntitlementsJSON string `json:"-"`
	Status           string `json:"status"`
	SortOrder        int    `json:"sort_order"`
}

type Subscription struct {
	orm.Model
	UserID                  uint       `json:"user_id"`
	PlanID                  uint       `json:"plan_id"`
	Status                  string     `json:"status"`
	StartsAt                *time.Time `json:"starts_at"`
	EndsAt                  *time.Time `json:"ends_at"`
	GracePeriodEndsAt       *time.Time `json:"grace_period_ends_at"`
	CanceledAt              *time.Time `json:"canceled_at"`
	EntitlementSnapshotJSON string     `json:"-"`
}

type SubscriptionNotificationDelivery struct {
	orm.Model
	SubscriptionID uint       `json:"subscription_id"`
	UserID         uint       `json:"user_id"`
	EventKey       string     `json:"event_key"`
	Kind           string     `json:"kind"`
	Channel        string     `json:"channel"`
	Status         string     `json:"status"`
	Attempts       int        `json:"attempts"`
	NextAttemptAt  *time.Time `json:"next_attempt_at"`
	SentAt         *time.Time `json:"sent_at"`
	LastError      string     `json:"last_error"`
}

type UsageLedger struct {
	orm.Model
	UserID         uint   `json:"user_id"`
	ResourceType   string `json:"resource_type"`
	Delta          int64  `json:"delta"`
	SourceType     string `json:"source_type"`
	SourceID       string `json:"source_id"`
	IdempotencyKey string `json:"-"`
	PeriodKey      string `json:"period_key"`
}
