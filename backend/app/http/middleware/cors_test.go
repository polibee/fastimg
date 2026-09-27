package middleware

import (
	"strings"
	"testing"
)

func TestCORSAllowsIdempotencyKeyForBrowserUploads(t *testing.T) {
	if !strings.Contains(corsAllowedHeaders, "Idempotency-Key") {
		t.Fatalf("browser upload preflight must allow Idempotency-Key, got %q", corsAllowedHeaders)
	}
}
