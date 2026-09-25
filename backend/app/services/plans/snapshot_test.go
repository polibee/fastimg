package planservices

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/app/services/quota"
)

func TestBuildSubscriptionSnapshotCapturesPlanTermsAndStableVersion(t *testing.T) {
	plan := models.Plan{
		Code:             "creator",
		Name:             "Creator",
		Description:      "For creators",
		PriceAmount:      1299,
		Currency:         "CNY",
		BillingPeriod:    "monthly",
		EntitlementsJSON: `{"storage_bytes":2000000000,"max_file_bytes":20000000,"daily_uploads":200,"monthly_api_uploads":1000,"monthly_bandwidth_bytes":10000000000,"transform_count":1000,"api_rate_per_minute":60,"token_limit":3,"ads_enabled":false}`,
	}

	snapshot, err := BuildSubscriptionSnapshot(plan)
	if err != nil {
		t.Fatalf("BuildSubscriptionSnapshot() error = %v", err)
	}
	parsed, err := ParseSubscriptionSnapshot(snapshot)
	if err != nil {
		t.Fatalf("ParseSubscriptionSnapshot() error = %v", err)
	}
	if !parsed.HasPlan || parsed.Plan.Name != plan.Name || parsed.Plan.PriceAmount != plan.PriceAmount {
		t.Fatalf("snapshot plan = %+v, hasPlan = %v", parsed.Plan, parsed.HasPlan)
	}
	if parsed.PlanVersion == "" || !strings.HasPrefix(parsed.PlanVersion, "sha256:") {
		t.Fatalf("plan version = %q, want content-addressed sha256 version", parsed.PlanVersion)
	}
	if _, err := quota.ParseEntitlementJSON(snapshot); err != nil {
		t.Fatalf("quota reader must accept versioned snapshot: %v", err)
	}

	plan.Description = "Edited after subscription"
	plan.PriceAmount = 1999
	updated, err := BuildSubscriptionSnapshot(plan)
	if err != nil {
		t.Fatalf("BuildSubscriptionSnapshot(updated) error = %v", err)
	}
	if updated == snapshot {
		t.Fatal("editing plan terms must create a distinct immutable snapshot version")
	}
	if parsed.Plan.Description != "For creators" || parsed.Plan.PriceAmount != 1299 {
		t.Fatalf("previous snapshot changed after plan edit: %+v", parsed.Plan)
	}
}

func TestParseSubscriptionSnapshotAcceptsLegacyEntitlementJSON(t *testing.T) {
	legacy, err := json.Marshal(quota.DefaultFreeEntitlement())
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseSubscriptionSnapshot(string(legacy))
	if err != nil {
		t.Fatalf("ParseSubscriptionSnapshot(legacy) error = %v", err)
	}
	if parsed.HasPlan || parsed.Entitlements.StorageBytes != quota.DefaultFreeEntitlement().StorageBytes {
		t.Fatalf("legacy snapshot parsed as %+v", parsed)
	}
}

func TestParseSubscriptionSnapshotAcceptsVersionHashWrittenBeforeWatermarkField(t *testing.T) {
	plan := PlanSnapshot{ID: 1, Code: "free", Name: "Free", Description: "Free media hosting baseline", Currency: "CNY", BillingPeriod: "monthly"}
	type legacyEntitlements struct {
		StorageBytes          int64 `json:"storage_bytes"`
		MaxFileBytes          int64 `json:"max_file_bytes"`
		DailyUploads          int64 `json:"daily_uploads"`
		MonthlyAPIUploads     int64 `json:"monthly_api_uploads"`
		MonthlyBandwidthBytes int64 `json:"monthly_bandwidth_bytes"`
		TransformCount        int64 `json:"transform_count"`
		APIRatePerMinute      int64 `json:"api_rate_per_minute"`
		TokenLimit            int64 `json:"token_limit"`
		AdsEnabled            bool  `json:"ads_enabled"`
	}
	entitlements := legacyEntitlements{
		StorageBytes: 1000, MaxFileBytes: 100, DailyUploads: 10,
		MonthlyAPIUploads: 20, MonthlyBandwidthBytes: 3000, TransformCount: 40,
		APIRatePerMinute: 5, TokenLimit: 2, AdsEnabled: true,
	}
	versionInput, err := json.Marshal(struct {
		Plan         PlanSnapshot `json:"plan"`
		Entitlements any          `json:"entitlements"`
	}{Plan: plan, Entitlements: entitlements})
	require.NoError(t, err)
	digest := sha256.Sum256(versionInput)
	payload, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"plan_version":   fmt.Sprintf("sha256:%x", digest),
		"plan":           plan,
		"entitlements":   entitlements,
	})
	require.NoError(t, err)

	parsed, err := ParseSubscriptionSnapshot(string(payload))
	require.NoError(t, err)
	require.True(t, parsed.HasPlan)
	require.False(t, parsed.Entitlements.WatermarkEnabled)
}

func TestParseSubscriptionSnapshotRejectsTermsChangedWithoutVersionChange(t *testing.T) {
	encoded, err := BuildSubscriptionSnapshot(models.Plan{
		Code: "creator", Name: "Creator", PriceAmount: 1299, Currency: "CNY", BillingPeriod: "monthly",
		EntitlementsJSON: `{"storage_bytes":2000000000,"max_file_bytes":20000000,"daily_uploads":200,"monthly_api_uploads":1000,"monthly_bandwidth_bytes":10000000000,"transform_count":1000,"api_rate_per_minute":60,"token_limit":3,"ads_enabled":false}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	var changed map[string]any
	if err := json.Unmarshal([]byte(encoded), &changed); err != nil {
		t.Fatal(err)
	}
	changed["plan"].(map[string]any)["price_amount"] = float64(1)
	tampered, err := json.Marshal(changed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseSubscriptionSnapshot(string(tampered)); err == nil {
		t.Fatal("snapshot terms changed without a matching plan version, want error")
	}
}
