package links

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeHotlinkDomainRemovesSchemeAndTrailingDot(t *testing.T) {
	got, err := normalizeHotlinkDomain(" HTTPS://Example.COM:443/path ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "example.com:443" {
		t.Fatalf("normalized domain = %q", got)
	}
	if _, err := normalizeHotlinkDomain("javascript:alert(1)"); err == nil {
		t.Fatal("unsafe domain scheme should be rejected")
	}
}

func TestSignedMediaURLRejectsTamperingAndExpiry(t *testing.T) {
	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	signature := signMediaURL(42, "thumbnail", expiresAt, "development-secret")
	if !verifyMediaURL(42, "thumbnail", expiresAt, signature, "development-secret", time.Now().UTC()) {
		t.Fatal("valid signature should verify")
	}
	if verifyMediaURL(43, "thumbnail", expiresAt, signature, "development-secret", time.Now().UTC()) {
		t.Fatal("media id changes must invalidate the signature")
	}
	if verifyMediaURL(42, "original", expiresAt, signature, "development-secret", time.Now().UTC()) {
		t.Fatal("variant changes must invalidate the signature")
	}
	if verifyMediaURL(42, "thumbnail", expiresAt, signature, "development-secret", expiresAt.Add(time.Second)) {
		t.Fatal("expired signatures must be rejected")
	}
	if verifyMediaURL(42, "thumbnail", expiresAt, strings.TrimSuffix(signature, "a")+"b", "development-secret", time.Now().UTC()) {
		t.Fatal("signature changes must be rejected")
	}
}

func TestStableMediaURLSignature(t *testing.T) {
	const secret = "test-app-key"
	signature := signStableMediaURL(42, "original", secret)
	if signature == "" {
		t.Fatal("stable signature must not be empty")
	}
	if !verifyStableMediaURL(42, "original", signature, secret) {
		t.Fatal("stable signature should verify")
	}
	if verifyStableMediaURL(42, "thumbnail", signature, secret) {
		t.Fatal("stable signature must be bound to its variant")
	}
	if verifyStableMediaURL(43, "original", signature, secret) {
		t.Fatal("stable signature must be bound to its media")
	}
	if verifyStableMediaURL(42, "original", signature, "other-key") {
		t.Fatal("stable signature must be bound to the app key")
	}
}

func TestHotlinkPolicyModes(t *testing.T) {
	domains := []string{"example.com", "images.example.org:8443"}
	checks := []struct {
		name           string
		mode           string
		referer        string
		allowNoReferer bool
		signatureValid bool
		want           bool
	}{
		{name: "off allows", mode: HotlinkModeOff, referer: "https://elsewhere.test/page", want: true},
		{name: "referer allows registered host", mode: HotlinkModeReferer, referer: "https://EXAMPLE.com/article", want: true},
		{name: "referer rejects unknown host", mode: HotlinkModeReferer, referer: "https://elsewhere.test/page", want: false},
		{name: "referer rejects missing by default", mode: HotlinkModeReferer, want: false},
		{name: "referer can allow missing", mode: HotlinkModeReferer, allowNoReferer: true, want: true},
		{name: "signed requires signature", mode: HotlinkModeSigned, referer: "https://example.com", want: false},
		{name: "signed accepts signature", mode: HotlinkModeSigned, signatureValid: true, want: true},
		{name: "hybrid accepts registered host", mode: HotlinkModeHybrid, referer: "https://images.example.org:8443/page", want: true},
		{name: "hybrid accepts signature", mode: HotlinkModeHybrid, signatureValid: true, want: true},
		{name: "hybrid rejects both", mode: HotlinkModeHybrid, referer: "https://elsewhere.test", want: false},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if got := allowsHotlink(check.mode, check.referer, check.allowNoReferer, domains, check.signatureValid); got != check.want {
				t.Fatalf("allowsHotlink() = %v, want %v", got, check.want)
			}
		})
	}
}
