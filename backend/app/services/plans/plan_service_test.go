package planservices

import (
	"encoding/json"
	"testing"

	"goravel/app/services/quota"
)

func TestFreeEntitlementSnapshotCanBeSerialized(t *testing.T) {
	payload, err := json.Marshal(quota.DefaultFreeEntitlement())
	if err != nil {
		t.Fatalf("marshal free entitlement: %v", err)
	}
	if len(payload) == 0 {
		t.Fatal("expected non-empty entitlement snapshot")
	}
}
