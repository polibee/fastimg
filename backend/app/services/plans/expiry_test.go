package planservices

import (
	"reflect"
	"testing"
	"time"
)

func TestNormalizeExpiryReminderDays(t *testing.T) {
	got := normalizeExpiryReminderDays("1, 7, 3, 3, 0, -1, invalid")
	want := []int{1, 3, 7}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeExpiryReminderDays() = %#v, want %#v", got, want)
	}
}

func TestClassifySubscriptionExpiry(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	past := now.Add(-time.Minute)

	tests := []struct {
		name   string
		endsAt *time.Time
		want   string
	}{
		{name: "manual subscription without end date remains active", endsAt: nil, want: subscriptionStateActive},
		{name: "future end date remains active", endsAt: &future, want: subscriptionStateActive},
		{name: "past end date is expired", endsAt: &past, want: subscriptionStateExpired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifySubscriptionExpiry(tt.endsAt, now); got != tt.want {
				t.Fatalf("classifySubscriptionExpiry() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGracePeriodEndsAt(t *testing.T) {
	expiresAt := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	want := expiresAt.Add(72 * time.Hour)
	if got := gracePeriodEndsAt(expiresAt, 3); !got.Equal(want) {
		t.Fatalf("gracePeriodEndsAt() = %s, want %s", got, want)
	}
}

func TestExpiryNotificationKey(t *testing.T) {
	if got, want := expiryNotificationKey(42, "7d"), "subscription:42:expiry:7d"; got != want {
		t.Fatalf("expiryNotificationKey() = %q, want %q", got, want)
	}
}

func TestSubscriptionTermEndsAt(t *testing.T) {
	startsAt := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		period    string
		trialDays int
		want      time.Time
	}{
		{period: "monthly", want: time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)},
		{period: "yearly", trialDays: 7, want: time.Date(2027, 10, 3, 12, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		got, err := subscriptionTermEndsAt(startsAt, tt.period, tt.trialDays)
		if err != nil {
			t.Fatalf("subscriptionTermEndsAt(%q) error = %v", tt.period, err)
		}
		if !got.Equal(tt.want) {
			t.Fatalf("subscriptionTermEndsAt(%q) = %s, want %s", tt.period, got, tt.want)
		}
	}
}
