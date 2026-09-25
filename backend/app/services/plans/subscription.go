package planservices

import (
	"errors"
	"fmt"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
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
	return &subscription, true, nil
}

func (s querySubscriptionStore) EnsureFree(userID uint) error {
	return EnsureFreeSubscriptionWithQuery(s.query, userID)
}
