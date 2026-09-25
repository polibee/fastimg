package planservices

import (
	"errors"
	"testing"

	"goravel/app/models"
)

type fakeSubscriptionStore struct {
	active      *models.Subscription
	ensureCalls int
	ensureErr   error
}

func (s *fakeSubscriptionStore) FindActive(uint) (*models.Subscription, bool, error) {
	if s.active == nil {
		return nil, false, nil
	}
	return s.active, true, nil
}

func (s *fakeSubscriptionStore) EnsureFree(userID uint) error {
	s.ensureCalls++
	if s.ensureErr != nil {
		return s.ensureErr
	}
	s.active = &models.Subscription{UserID: userID, Status: "active"}
	return nil
}

func TestEnsureActiveSubscriptionRepairsMissingFreeSubscription(t *testing.T) {
	store := &fakeSubscriptionStore{}
	subscription, err := ensureActiveSubscription(store, 27)
	if err != nil {
		t.Fatalf("ensureActiveSubscription() error = %v", err)
	}
	if subscription == nil || subscription.UserID != 27 || store.ensureCalls != 1 {
		t.Fatalf("subscription=%+v ensureCalls=%d, want active user 27 and one free-plan provision", subscription, store.ensureCalls)
	}
}

func TestEnsureActiveSubscriptionDoesNotMaskProvisioningError(t *testing.T) {
	store := &fakeSubscriptionStore{ensureErr: errors.New("free plan missing")}
	if _, err := ensureActiveSubscription(store, 27); !errors.Is(err, store.ensureErr) {
		t.Fatalf("ensureActiveSubscription() error = %v, want provisioning error", err)
	}
}

func TestEnsureActiveSubscriptionDoesNotRewriteExistingSubscription(t *testing.T) {
	want := &models.Subscription{UserID: 27, PlanID: 3, Status: "active"}
	store := &fakeSubscriptionStore{active: want}
	got, err := ensureActiveSubscription(store, 27)
	if err != nil || got != want || store.ensureCalls != 0 {
		t.Fatalf("ensureActiveSubscription() = (%+v, %v), ensureCalls=%d; want existing subscription unchanged", got, err, store.ensureCalls)
	}
}
