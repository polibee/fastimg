package billing

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/app/services/quota"
)

func TestBuildPlanPriceSnapshotPreservesImmutableCommercialTerms(t *testing.T) {
	plan := models.Plan{
		Code: "creator", Name: "Creator", Description: "Creator plan",
		Currency: "USD", BillingPeriod: "monthly",
		EntitlementsJSON: `{"storage_bytes":2000000000,"max_file_bytes":20000000,"daily_uploads":500,"monthly_api_uploads":5000,"monthly_bandwidth_bytes":20000000000,"transform_count":2000,"api_rate_per_minute":120,"token_limit":3,"ads_enabled":false}`,
	}
	plan.ID = 7
	price := models.PlanPrice{
		PlanID: plan.ID, Version: "creator-usd-monthly-v1", Currency: "USD",
		AmountMinor: 990, BillingPeriod: "monthly", TrialDays: 0, Status: "active",
	}
	price.ID = 21

	snapshot, err := BuildPlanPriceSnapshot(plan, price)
	require.NoError(t, err)
	require.Equal(t, uint(7), snapshot.PlanID)
	require.Equal(t, uint(21), snapshot.PriceID)
	require.Equal(t, int64(990), snapshot.AmountMinor)
	require.Equal(t, "USD", snapshot.Currency)
	require.Equal(t, quota.Entitlement{StorageBytes: 2000000000, MaxFileBytes: 20000000, DailyUploads: 500, MonthlyAPIUploads: 5000, MonthlyBandwidthBytes: 20000000000, TransformCount: 2000, APIRatePerMinute: 120, TokenLimit: 3, AdsEnabled: false}, snapshot.Entitlements)
}

func TestValidatePlanPriceRejectsInvalidCommercialTerms(t *testing.T) {
	cases := []struct {
		name  string
		price models.PlanPrice
	}{
		{name: "negative amount", price: models.PlanPrice{AmountMinor: -1, Currency: "USD", BillingPeriod: "monthly", Status: "active"}},
		{name: "lowercase currency", price: models.PlanPrice{AmountMinor: 100, Currency: "usd", BillingPeriod: "monthly", Status: "active"}},
		{name: "unsupported period", price: models.PlanPrice{AmountMinor: 100, Currency: "USD", BillingPeriod: "weekly", Status: "active"}},
		{name: "unsupported status", price: models.PlanPrice{AmountMinor: 100, Currency: "USD", BillingPeriod: "monthly", Status: "published"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Error(t, ValidatePlanPrice(tc.price))
		})
	}
}

func TestPlanPriceSnapshotDoesNotReadMutablePlanAfterCreation(t *testing.T) {
	plan := models.Plan{Code: "pro", Name: "Pro", Currency: "USD", BillingPeriod: "yearly", EntitlementsJSON: `{"storage_bytes":100,"max_file_bytes":10,"daily_uploads":1,"monthly_api_uploads":2,"monthly_bandwidth_bytes":3,"transform_count":4,"api_rate_per_minute":5,"token_limit":6,"ads_enabled":false}`}
	plan.ID = 2
	price := models.PlanPrice{PlanID: 2, Version: "v1", Currency: "USD", AmountMinor: 12000, BillingPeriod: "yearly", Status: "active"}
	price.ID = 3
	snapshot, err := BuildPlanPriceSnapshot(plan, price)
	require.NoError(t, err)

	plan.Name = "Changed Pro"
	plan.Currency = "EUR"
	plan.EntitlementsJSON = `{"storage_bytes":999,"max_file_bytes":10,"daily_uploads":1,"monthly_api_uploads":2,"monthly_bandwidth_bytes":3,"transform_count":4,"api_rate_per_minute":5,"token_limit":6,"ads_enabled":true}`

	require.Equal(t, "Pro", snapshot.PlanName)
	require.Equal(t, "USD", snapshot.Currency)
	require.Equal(t, int64(100), snapshot.Entitlements.StorageBytes)
}
