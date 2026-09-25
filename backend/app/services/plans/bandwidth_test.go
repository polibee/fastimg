package planservices

import (
	"testing"
	"time"

	"goravel/app/services/quota"
)

func TestBandwidthUsageRecordUsesCurrentUTCMonthAndDownloadSource(t *testing.T) {
	period := bandwidthUsagePeriod(time.Date(2026, 9, 25, 23, 30, 0, 0, time.FixedZone("CST", 8*60*60)))
	if period != "2026-09" {
		t.Fatalf("bandwidth period = %q, want 2026-09", period)
	}
	record := quota.UsageRecord{UserID: 7, ResourceType: "bandwidth", Delta: 2048, SourceType: "download", SourceID: "access:19", PeriodKey: period}
	if err := validateBandwidthUsageRecord(record); err != nil {
		t.Fatalf("validate bandwidth record: %v", err)
	}
}
