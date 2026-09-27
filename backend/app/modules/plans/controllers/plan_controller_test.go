package controllers

import (
	"encoding/json"
	"strings"
	"testing"

	"goravel/app/models"
)

func TestPublicPlanUsesStableEntitlementFieldNames(t *testing.T) {
	public, err := publicPlan(models.Plan{
		Code: "creator", Name: "Creator", PriceAmount: 499, Currency: "CNY",
		BillingPeriod: "monthly", Status: "active",
		EntitlementsJSON: `{"StorageBytes":1000,"MaxFileBytes":100,"DailyUploads":10,"MonthlyAPIUploads":20,"MonthlyBandwidthBytes":3000,"TransformCount":40,"APIRatePerMinute":5,"TokenLimit":2,"AdsEnabled":true}`,
	})
	if err != nil {
		t.Fatalf("build public plan: %v", err)
	}
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("marshal public plan: %v", err)
	}
	if !strings.Contains(string(encoded), `"storage_bytes":1000`) || strings.Contains(string(encoded), `"StorageBytes"`) {
		t.Fatalf("public plan entitlement contract is not canonical: %s", encoded)
	}
	if strings.Contains(string(encoded), `"price_amount"`) || strings.Contains(string(encoded), `"billing_period"`) || strings.Contains(string(encoded), `"currency"`) {
		t.Fatalf("public plan must not expose legacy plan-level price fields: %s", encoded)
	}
}

func TestPublicPlanRejectsMalformedEntitlements(t *testing.T) {
	if _, err := publicPlan(models.Plan{EntitlementsJSON: `{}`}); err == nil {
		t.Fatal("malformed plan entitlements were exposed as a public plan")
	}
}
