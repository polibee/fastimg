package middleware

import (
	"strings"
	"testing"
)

func TestNormalizeRequestIDKeepsSafeClientID(t *testing.T) {
	if got := normalizeRequestID("client.req-123"); got != "client.req-123" {
		t.Fatalf("expected safe client request id to be preserved, got %q", got)
	}
}

func TestNormalizeRequestIDReplacesUnsafeOrOversizedID(t *testing.T) {
	for _, input := range []string{"bad\r\nheader", strings.Repeat("x", 65)} {
		got := normalizeRequestID(input)
		if got == input || !strings.HasPrefix(got, "req_") {
			t.Fatalf("expected generated request id for %q, got %q", input, got)
		}
	}
}

func TestRequestIDIsRetryableForTransientStatuses(t *testing.T) {
	for _, status := range []int{429, 500, 502, 503, 504} {
		if !retryableStatus(status) {
			t.Fatalf("expected status %d to be retryable", status)
		}
	}
	for _, status := range []int{400, 401, 403, 404, 409, 422} {
		if retryableStatus(status) {
			t.Fatalf("expected status %d not to be retryable", status)
		}
	}
}
