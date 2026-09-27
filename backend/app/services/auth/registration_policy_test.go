package authservices

import "testing"

func TestParseWhitelistDomainsNormalizesAndDeduplicates(t *testing.T) {
	got := ParseWhitelistDomains(" Example.com, example.com\nsub.example.com; ")
	want := []string{"example.com", "sub.example.com"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestEmailAllowedByWhitelist(t *testing.T) {
	if !EmailAllowedByWhitelist("person@Example.com", []string{"example.com"}) {
		t.Fatal("expected email domain to match case-insensitively")
	}
	if EmailAllowedByWhitelist("person@other.com", []string{"example.com"}) {
		t.Fatal("unexpected non-whitelisted email match")
	}
}

func TestVerificationExpiryIsBounded(t *testing.T) {
	if got := NormalizeVerificationExpiry(0); got != defaultVerificationExpiryMinutes {
		t.Fatalf("expected default expiry, got %d", got)
	}
	if got := NormalizeVerificationExpiry(1); got != minVerificationExpiryMinutes {
		t.Fatalf("expected minimum expiry, got %d", got)
	}
	if got := NormalizeVerificationExpiry(99999); got != maxVerificationExpiryMinutes {
		t.Fatalf("expected maximum expiry, got %d", got)
	}
}
