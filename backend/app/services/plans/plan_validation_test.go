package planservices

import (
	"errors"
	"testing"
)

func validPlanPayload() map[string]any {
	return map[string]any{
		"code": "creator", "name": "Creator", "price_amount": float64(499),
		"currency": "CNY", "billing_period": "monthly", "status": "active",
		"entitlements_json": `{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true}`,
	}
}

func TestValidatePlanWriteAcceptsValidPlan(t *testing.T) {
	if err := PreparePlanWrite("create", validPlanPayload()); err != nil {
		t.Fatalf("valid plan rejected: %v", err)
	}
}

func TestValidatePlanWriteRejectsInvalidDomainValues(t *testing.T) {
	cases := []struct {
		name  string
		field string
		value any
	}{
		{name: "negative price", field: "price_amount", value: float64(-1)},
		{name: "fractional price", field: "price_amount", value: float64(1.5)},
		{name: "invalid currency", field: "currency", value: "cny"},
		{name: "unknown billing period", field: "billing_period", value: "weekly"},
		{name: "unknown status", field: "status", value: "pending"},
		{name: "malformed entitlement JSON", field: "entitlements_json", value: `{`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			payload := validPlanPayload()
			payload[testCase.field] = testCase.value
			if err := PreparePlanWrite("update", payload); !errors.Is(err, ErrInvalidPlan) {
				t.Fatalf("error = %v, want ErrInvalidPlan", err)
			}
		})
	}
}

func TestValidatePlanWriteAllowsPartialUpdate(t *testing.T) {
	if err := PreparePlanWrite("update", map[string]any{"status": "disabled"}); err != nil {
		t.Fatalf("valid partial update rejected: %v", err)
	}
}

func TestPreparePlanWriteCanonicalizesPriceAndLegacyEntitlements(t *testing.T) {
	payload := validPlanPayload()
	payload["price_amount"] = float64(499)
	payload["entitlements_json"] = `{"StorageBytes":1000,"MaxFileBytes":100,"DailyUploads":10,"MonthlyAPIUploads":20,"MonthlyBandwidthBytes":3000,"TransformCount":40,"APIRatePerMinute":5,"TokenLimit":2,"AdsEnabled":true}`

	if err := PreparePlanWrite("create", payload); err != nil {
		t.Fatalf("prepare valid legacy payload: %v", err)
	}
	if payload["price_amount"] != int64(499) {
		t.Fatalf("price_amount = %#v (%T), want int64(499)", payload["price_amount"], payload["price_amount"])
	}
	wantEntitlements := `{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true,"watermark_enabled":false}`
	if payload["entitlements_json"] != wantEntitlements {
		t.Fatalf("entitlements_json = %v, want %s", payload["entitlements_json"], wantEntitlements)
	}
}
