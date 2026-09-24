package seeders

import (
	"encoding/json"
	"reflect"
	"testing"

	"goravel/app/services/quota"
)

func TestBuildFastImgFreePlanUsesDefaultEntitlements(t *testing.T) {
	plan, err := buildFastImgFreePlan()
	if err != nil {
		t.Fatalf("build free plan: %v", err)
	}
	if plan.Code != "free" || plan.PriceAmount != 0 || plan.Status != "active" {
		t.Fatalf("unexpected free plan identity or status: %+v", plan)
	}

	var got any
	if err := json.Unmarshal([]byte(plan.EntitlementsJSON), &got); err != nil {
		t.Fatalf("decode free plan entitlements: %v", err)
	}
	wantJSON, err := json.Marshal(quota.DefaultFreeEntitlement())
	if err != nil {
		t.Fatalf("encode default free entitlements: %v", err)
	}
	var want any
	if err := json.Unmarshal(wantJSON, &want); err != nil {
		t.Fatalf("decode default free entitlements: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("free plan entitlements = %v, want %v", got, want)
	}
}

func TestBuildFastImgPlansIncludesFreeAndPaidTiers(t *testing.T) {
	plans, err := buildFastImgPlans()
	if err != nil {
		t.Fatalf("build plans: %v", err)
	}
	if len(plans) != 3 || plans[0].Code != "free" || plans[1].Code != "creator" || plans[2].Code != "pro" {
		t.Fatalf("unexpected plan catalog: %+v", plans)
	}
	if plans[0].PriceAmount != 0 || plans[1].PriceAmount <= 0 || plans[2].PriceAmount <= plans[1].PriceAmount {
		t.Fatalf("unexpected plan prices: %+v", plans)
	}
	for _, plan := range plans {
		if _, err := quota.ParseEntitlementJSON(plan.EntitlementsJSON); err != nil {
			t.Fatalf("plan %s has invalid entitlements: %v", plan.Code, err)
		}
	}
}
