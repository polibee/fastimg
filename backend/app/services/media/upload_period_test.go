package media

import (
	"testing"
	"time"
)

func TestUploadUsagePeriodUsesReservationMonthAcrossMonthBoundary(t *testing.T) {
	reservedAt := time.Date(2026, time.September, 30, 23, 59, 0, 0, time.UTC)
	completedAt := time.Date(2026, time.October, 1, 0, 1, 0, 0, time.UTC)

	if got := uploadUsagePeriodKey(reservedAt, completedAt); got != "2026-09" {
		t.Fatalf("upload usage period = %q, want reservation month 2026-09", got)
	}
}

func TestUploadUsagePeriodFallsBackToCompletionWhenReservationTimeIsMissing(t *testing.T) {
	completedAt := time.Date(2026, time.October, 1, 0, 1, 0, 0, time.UTC)

	if got := uploadUsagePeriodKey(time.Time{}, completedAt); got != "2026-10" {
		t.Fatalf("upload usage period = %q, want completion month 2026-10", got)
	}
}
