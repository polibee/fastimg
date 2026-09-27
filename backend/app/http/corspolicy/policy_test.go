package corspolicy

import "testing"

func TestConfiguredOriginsAreExactAndNeverWildcard(t *testing.T) {
	if !AllowedOriginFor("https://fastimg.example", "production", "https://fastimg.example, https://admin.fastimg.example") {
		t.Fatal("configured production origin should be allowed")
	}
	if AllowedOriginFor("https://evil.example", "production", "https://fastimg.example") {
		t.Fatal("unconfigured production origin must be rejected")
	}
	if AllowedOriginFor("https://evil.example", "production", "*") {
		t.Fatal("wildcard origin must never be accepted")
	}
}

func TestDevelopmentOriginsRemainAvailableWithoutConfiguration(t *testing.T) {
	if !AllowedOriginFor("http://127.0.0.1:53083", "local", "") {
		t.Fatal("local development origin should remain available")
	}
	if AllowedOriginFor("http://127.0.0.1:53083", "production", "") {
		t.Fatal("production must not inherit development origins")
	}
}

func TestAllowedOrigin(t *testing.T) {
	if !AllowedOrigin("http://127.0.0.1:4174") {
		t.Fatal("expected the local admin preview origin to be allowed")
	}
	if !AllowedOrigin("http://127.0.0.1:5182") {
		t.Fatal("expected the isolated local admin verification origin to be allowed")
	}
	if !AllowedOrigin("http://127.0.0.1:53083") {
		t.Fatal("expected the isolated WSL preview origin to be allowed")
	}
	if !AllowedOrigin("http://127.0.0.1:53084") {
		t.Fatal("expected the FastImg frontend origin to be allowed")
	}
	if AllowedOrigin("https://example.com") {
		t.Fatal("unexpectedly allowed an untrusted origin")
	}
}
