package developer

import (
	"testing"
	"time"
)

func TestRateLimitDecisionAllowsWithinWindowAndReportsRemaining(t *testing.T) {
	now := time.Date(2026, 9, 25, 0, 0, 30, 0, time.UTC)
	allowed, remaining, retryAfter, resetAt := rateLimitDecision(2, 3, now.Add(-10*time.Second), now, time.Minute)

	if !allowed {
		t.Fatal("expected request within limit to be allowed")
	}
	if remaining != 0 {
		t.Fatalf("expected no remaining requests after third attempt, got %d", remaining)
	}
	if retryAfter != 0 {
		t.Fatalf("expected no retry delay for an allowed request, got %d", retryAfter)
	}
	if !resetAt.Equal(now.Add(50 * time.Second)) {
		t.Fatalf("unexpected reset time: %s", resetAt)
	}
}

func TestRateLimitDecisionResetsExpiredWindow(t *testing.T) {
	now := time.Date(2026, 9, 25, 0, 2, 0, 0, time.UTC)
	allowed, remaining, retryAfter, resetAt := rateLimitDecision(99, 100, now.Add(-2*time.Minute), now, time.Minute)

	if !allowed || remaining != 99 || retryAfter != 0 {
		t.Fatalf("expected expired window to reset, allowed=%v remaining=%d retry_after=%d", allowed, remaining, retryAfter)
	}
	if !resetAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("unexpected reset time after reset: %s", resetAt)
	}
}

func TestRateLimitDecisionRejectsAtLimit(t *testing.T) {
	now := time.Date(2026, 9, 25, 0, 3, 10, 0, time.UTC)
	allowed, remaining, retryAfter, resetAt := rateLimitDecision(3, 3, now.Add(-10*time.Second), now, time.Minute)

	if allowed || remaining != 0 {
		t.Fatalf("expected limit rejection, allowed=%v remaining=%d", allowed, remaining)
	}
	if retryAfter != 50 {
		t.Fatalf("expected 50 second retry delay, got %d", retryAfter)
	}
	if !resetAt.Equal(now.Add(50 * time.Second)) {
		t.Fatalf("unexpected rejection reset time: %s", resetAt)
	}
}
