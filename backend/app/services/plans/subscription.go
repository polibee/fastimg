package planservices

import (
	"errors"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	settingsservices "goravel/app/services/settings"
)

var ErrInvalidSubscriptionOwner = errors.New("subscription owner is required")

type subscriptionStore interface {
	FindActive(userID uint) (*models.Subscription, bool, error)
	EnsureFree(userID uint) error
}

func ensureActiveSubscription(store subscriptionStore, userID uint) (*models.Subscription, error) {
	if store == nil || userID == 0 {
		return nil, ErrInvalidSubscriptionOwner
	}
	active, found, err := store.FindActive(userID)
	if err != nil {
		return nil, err
	}
	if found {
		return active, nil
	}
	if err := store.EnsureFree(userID); err != nil {
		return nil, fmt.Errorf("provision Free subscription: %w", err)
	}
	active, found, err = store.FindActive(userID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrSubscriptionUnavailable
	}
	return active, nil
}

// EnsureActiveSubscriptionWithQuery resolves the active subscription from the
// provided query/transaction and provisions Free only when no active row exists.
func EnsureActiveSubscriptionWithQuery(query orm.Query, userID uint) (*models.Subscription, error) {
	return ensureActiveSubscription(querySubscriptionStore{query: query}, userID)
}

type querySubscriptionStore struct{ query orm.Query }

// CreateSubscriptionWithQuery writes only columns known to the current
// installation. This keeps older development schemas readable until the
// expiry lifecycle migration is applied.
func CreateSubscriptionWithQuery(query orm.Query, subscription models.Subscription) error {
	now := time.Now().UTC()
	values := map[string]any{
		"user_id":                   subscription.UserID,
		"plan_id":                   subscription.PlanID,
		"status":                    subscription.Status,
		"starts_at":                 subscription.StartsAt,
		"ends_at":                   subscription.EndsAt,
		"canceled_at":               subscription.CanceledAt,
		"entitlement_snapshot_json": subscription.EntitlementSnapshotJSON,
		"created_at":                now,
		"updated_at":                now,
	}
	if facades.Schema().HasColumn("subscriptions", "grace_period_ends_at") {
		values["grace_period_ends_at"] = subscription.GracePeriodEndsAt
	}
	return query.Table("subscriptions").Create(values)
}

func (s querySubscriptionStore) FindActive(userID uint) (*models.Subscription, bool, error) {
	query := s.query.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, "active")
	exists, err := query.Exists()
	if err != nil || !exists {
		return nil, exists, err
	}
	var subscription models.Subscription
	if err := query.First(&subscription); err != nil {
		return nil, false, err
	}
	if subscription.EndsAt != nil && !subscription.EndsAt.After(time.Now().UTC()) {
		policy := expiryPolicy(settingsservices.NewSettingService())
		now := time.Now().UTC()
		graceEnds := gracePeriodEndsAt(subscription.EndsAt.UTC(), policy.GracePeriodDays)
		updates := map[string]any{"status": "expired", "updated_at": now}
		if facades.Schema().HasColumn("subscriptions", "grace_period_ends_at") {
			updates["grace_period_ends_at"] = graceEnds
		}
		if _, err := s.query.Where("id = ?", subscription.ID).Update(updates); err != nil {
			return nil, false, err
		}
		if err := EnsureFreeSubscriptionWithQuery(s.query, userID); err != nil {
			return nil, false, err
		}
		if err := s.query.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, "active").First(&subscription); err != nil {
			return nil, false, err
		}
	}
	return &subscription, true, nil
}

func (s querySubscriptionStore) EnsureFree(userID uint) error {
	return EnsureFreeSubscriptionWithQuery(s.query, userID)
}
