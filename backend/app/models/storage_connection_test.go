package models

import "testing"

func TestStorageConnectionProviderCodes(t *testing.T) {
	for _, code := range []string{"local", "cloudflare_r2", "aliyun_oss", "tencent_cos"} {
		if !IsStorageProviderCode(code) {
			t.Errorf("IsStorageProviderCode(%q) = false", code)
		}
	}
	if IsStorageProviderCode("unknown") {
		t.Fatal("unknown provider must be rejected")
	}
}

func TestStorageConnectionStatusValues(t *testing.T) {
	for _, status := range []string{"disabled", "incomplete", "healthy", "degraded", "error"} {
		if !IsStorageConnectionStatus(status) {
			t.Errorf("IsStorageConnectionStatus(%q) = false", status)
		}
	}
	if IsStorageConnectionStatus("ready") {
		t.Fatal("legacy ready status must not be accepted")
	}
}
