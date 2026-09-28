package services

import "testing"

func TestValidURLRejectsPrivateAndUnsafeTargets(t *testing.T) {
	for _, value := range []string{"javascript:alert(1)", "data:text/html,x", "http://localhost/site", "http://127.0.0.1/site", "http://10.0.0.4/site", "http://service.internal/site", "https://user:pass@example.com/site"} {
		if validURL(value, false) {
			t.Fatalf("validURL accepted unsafe target %q", value)
		}
	}
}

func TestValidURLAcceptsPublicHTTPSTargets(t *testing.T) {
	for _, value := range []string{"https://example.com", "http://cdn.example.org/logo.png"} {
		if !validURL(value, false) {
			t.Fatalf("validURL rejected public target %q", value)
		}
	}
}
