package quota

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestFreeEntitlementProvidesUsableBaseline(t *testing.T) {
	got := DefaultFreeEntitlement()

	if got.StorageBytes != 1_000_000_000 {
		t.Fatalf("storage = %d, want 1000000000", got.StorageBytes)
	}
	if got.MaxFileBytes != 10_000_000 {
		t.Fatalf("max file = %d, want 10000000", got.MaxFileBytes)
	}
	if got.DailyUploads != 100 || got.MonthlyAPIUploads != 500 {
		t.Fatalf("upload limits = %+v, want daily=100 monthly=500", got)
	}
	if !got.AdsEnabled {
		t.Fatal("free entitlement should allow ads")
	}
}

func TestReserveStorageRejectsLimitAndTracksReservation(t *testing.T) {
	state := UsageState{StorageBytes: 900, ReservedBytes: 50}
	limit := Entitlement{StorageBytes: 1_000, MaxFileBytes: 600}

	reservation, err := ReserveStorage(state, limit, 50)
	if err != nil {
		t.Fatalf("reserve storage: %v", err)
	}
	if reservation.Bytes != 50 || reservation.RemainingBytes != 0 {
		t.Fatalf("reservation = %+v", reservation)
	}

	if _, err := ReserveStorage(state, limit, 51); err != ErrStorageQuotaExceeded {
		t.Fatalf("expected storage quota error, got %v", err)
	}
}

func TestReserveStorageRejectsSingleFileLimit(t *testing.T) {
	_, err := ReserveStorage(UsageState{}, Entitlement{StorageBytes: 1_000, MaxFileBytes: 100}, 101)
	if err != ErrFileTooLarge {
		t.Fatalf("expected file too large, got %v", err)
	}
}

func TestReserveStorageRejectsNegativeUsageState(t *testing.T) {
	_, err := ReserveStorage(UsageState{StorageBytes: -1}, Entitlement{StorageBytes: 100}, 1)
	if !errors.Is(err, ErrStorageQuotaExceeded) {
		t.Fatalf("negative usage error = %v, want ErrStorageQuotaExceeded", err)
	}
}

func TestReleaseStorageNeverCreatesNegativeReservation(t *testing.T) {
	state := UsageState{StorageBytes: 100, ReservedBytes: 20}
	got := ReleaseStorage(state, 50)
	if got.StorageBytes != 50 || got.ReservedBytes != 0 {
		t.Fatalf("released state = %+v", got)
	}
}

func TestParseEntitlementJSONAcceptsCanonicalAndLegacySnapshots(t *testing.T) {
	want := Entitlement{
		StorageBytes: 1000, MaxFileBytes: 100, DailyUploads: 10,
		MonthlyAPIUploads: 20, MonthlyBandwidthBytes: 3000,
		TransformCount: 40, APIRatePerMinute: 5, TokenLimit: 2, AdsEnabled: true,
	}
	cases := []struct {
		name string
		json string
	}{
		{
			name: "snake case contract",
			json: `{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true}`,
		},
		{
			name: "legacy Go field names",
			json: `{"StorageBytes":1000,"MaxFileBytes":100,"DailyUploads":10,"MonthlyAPIUploads":20,"MonthlyBandwidthBytes":3000,"TransformCount":40,"APIRatePerMinute":5,"TokenLimit":2,"AdsEnabled":true}`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := ParseEntitlementJSON(testCase.json)
			if err != nil {
				t.Fatalf("parse entitlement: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("entitlement = %+v, want %+v", got, want)
			}
		})
	}
}

func TestParseEntitlementJSONReadsOptionalWatermarkEntitlement(t *testing.T) {
	payload := `{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true,"watermark_enabled":true}`

	got, err := ParseEntitlementJSON(payload)
	if err != nil {
		t.Fatalf("parse entitlement: %v", err)
	}
	if !got.WatermarkEnabled {
		t.Fatal("watermark entitlement should be enabled")
	}
}

func TestParseEntitlementJSONAcceptsVersionedSubscriptionSnapshot(t *testing.T) {
	payload := `{"schema_version":1,"plan_version":"sha256:abc","plan":{"code":"creator"},"entitlements":{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true}}`
	got, err := ParseEntitlementJSON(payload)
	if err != nil {
		t.Fatalf("parse versioned snapshot: %v", err)
	}
	if got.StorageBytes != 1000 || got.TokenLimit != 2 {
		t.Fatalf("entitlement = %+v", got)
	}
}

func TestAggregateUsageSeparatesLifetimeStorageFromCurrentPeriod(t *testing.T) {
	entries := []UsageEntry{
		{ResourceType: "storage_bytes", Delta: 300, PeriodKey: "2026-08"},
		{ResourceType: "storage", Delta: -100, PeriodKey: "2026-09"},
		{ResourceType: "api", Delta: 4, PeriodKey: "2026-08"},
		{ResourceType: "api", Delta: 7, PeriodKey: "2026-09"},
		{ResourceType: "bandwidth", Delta: 900, PeriodKey: "2026-09"},
	}
	got := AggregateUsage(entries, "2026-09")
	if got["storage"] != 200 || got["api"] != 7 || got["bandwidth"] != 900 {
		t.Fatalf("usage totals = %#v, want storage=200 api=7 bandwidth=900", got)
	}
}

func TestCheckUploadLimitAndBandwidthRespectUnlimitedAndExactBoundaries(t *testing.T) {
	if err := CheckUploadLimit(9, 10); err != nil {
		t.Fatalf("upload below limit: %v", err)
	}
	if err := CheckUploadLimit(10, 10); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("upload at limit error = %v, want ErrQuotaExceeded", err)
	}
	if err := CheckUploadLimit(10, 0); err != nil {
		t.Fatalf("zero upload limit should be unlimited: %v", err)
	}
	if err := CheckBandwidth(900, 100, 1000); err != nil {
		t.Fatalf("bandwidth exactly at limit: %v", err)
	}
	if err := CheckBandwidth(900, 101, 1000); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("bandwidth above limit error = %v, want ErrQuotaExceeded", err)
	}
	if err := CheckBandwidth(9223372036854775807, 1, 9223372036854775807); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("overflow-sized usage error = %v, want ErrQuotaExceeded", err)
	}
	if err := CheckBandwidth(0, -1, 1000); !errors.Is(err, ErrInvalidUsageAmount) {
		t.Fatalf("negative bandwidth error = %v, want ErrInvalidUsageAmount", err)
	}
}

func TestParseEntitlementJSONRejectsInvalidSnapshots(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{name: "malformed", json: `{`},
		{name: "unknown field", json: `{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true,"max_uploads":99}`},
		{name: "misspelled canonical field", json: `{"storage__bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true}`},
		{name: "missing field", json: `{"storage_bytes":1000,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2}`},
		{name: "negative limit", json: `{"storage_bytes":-1,"max_file_bytes":100,"daily_uploads":10,"monthly_api_uploads":20,"monthly_bandwidth_bytes":3000,"transform_count":40,"api_rate_per_minute":5,"token_limit":2,"ads_enabled":true}`},
		{name: "not an object", json: `null`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := ParseEntitlementJSON(testCase.json); !errors.Is(err, ErrInvalidEntitlement) {
				t.Fatalf("error = %v, want ErrInvalidEntitlement", err)
			}
		})
	}
}

func TestEntitlementJSONUsesStableSnakeCaseFields(t *testing.T) {
	encoded, err := json.Marshal(DefaultFreeEntitlement())
	if err != nil {
		t.Fatalf("marshal free entitlement: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("decode marshalled entitlement: %v", err)
	}
	if _, ok := fields["storage_bytes"]; !ok {
		t.Fatalf("canonical storage_bytes field missing from %s", encoded)
	}
	if _, ok := fields["StorageBytes"]; ok {
		t.Fatalf("legacy Go field name leaked into new JSON: %s", encoded)
	}
}
