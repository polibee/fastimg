package planservices

import (
	"encoding/json"
	"reflect"
	"testing"

	"goravel/app/services/quota"
)

func TestBuildDefaultFreePlanUsesCurrentFreeEntitlements(t *testing.T) {
	plan, err := buildDefaultFreePlan()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Code != "free" || plan.Status != "active" || plan.PriceAmount != 0 {
		t.Fatalf("unexpected default free plan: %+v", plan)
	}
	var got quota.Entitlement
	if err := json.Unmarshal([]byte(plan.EntitlementsJSON), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, quota.DefaultFreeEntitlement()) {
		t.Fatalf("default entitlements = %+v, want %+v", got, quota.DefaultFreeEntitlement())
	}
}
