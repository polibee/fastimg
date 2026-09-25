package feature

import (
	"errors"
	"testing"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
	"goravel/app/services/quota"
	"goravel/tests"
)

func TestEnsureActiveSubscriptionWithQueryProvisionsMissingFreeSnapshot(t *testing.T) {
	_ = tests.TestCase{}
	const userID uint = 4_000_000_001
	rollback := errors.New("rollback subscription test")

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		subscription, err := planservices.EnsureActiveSubscriptionWithQuery(tx, userID)
		if err != nil {
			return err
		}
		if subscription.UserID != userID || subscription.Status != "active" {
			t.Fatalf("provisioned subscription = %+v, want active subscription for test owner", subscription)
		}
		entitlement, err := quota.ParseEntitlementJSON(subscription.EntitlementSnapshotJSON)
		if err != nil {
			return err
		}
		if entitlement.StorageBytes <= 0 {
			t.Fatalf("provisioned Free storage limit = %d, want positive", entitlement.StorageBytes)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v, want intentional rollback after successful provision", err)
	}

	exists, err := facades.Orm().Query().Model(&models.Subscription{}).
		Where("user_id = ? AND status = ?", userID, "active").Exists()
	if err != nil {
		t.Fatalf("check rollback: %v", err)
	}
	if exists {
		t.Fatal("subscription test fixture persisted after rollback")
	}
}
