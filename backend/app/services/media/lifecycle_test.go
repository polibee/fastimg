package media

import (
	"testing"
	"time"
)

func TestMediaExpiryDecision(t *testing.T) {
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	if got := ExpiryState(now.Add(48*time.Hour), now, 72*time.Hour); got != ExpiryStateActive {
		t.Fatalf("expected active, got %s", got)
	}
	if got := ExpiryState(now.Add(12*time.Hour), now, 24*time.Hour); got != ExpiryStateExpiring {
		t.Fatalf("expected expiring, got %s", got)
	}
	if got := ExpiryState(now.Add(-time.Hour), now, 24*time.Hour); got != ExpiryStateExpired {
		t.Fatalf("expected expired, got %s", got)
	}
}
