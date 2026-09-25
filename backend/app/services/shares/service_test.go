package shares

import (
	"strings"
	"testing"
	"time"

	"goravel/app/models"
)

func TestShareTokenHashDoesNotExposeRawToken(t *testing.T) {
	raw, err := randomToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 64 || tokenHash(raw) == raw || len(tokenHash(raw)) != 64 {
		t.Fatalf("share token must be opaque and hashed: raw=%q hash=%q", raw, tokenHash(raw))
	}
}

func TestShareViewOnlyReturnsRawURLWhenCreated(t *testing.T) {
	link := models.ShareLink{MediaAssetID: 7, TokenPrefix: "abcdef12", Status: "active"}
	if got := shareView(link, "").URL; got != "" {
		t.Fatalf("list view must not reconstruct a share URL: %q", got)
	}
	if got := shareView(link, strings.Repeat("a", 64)).URL; got != "http://127.0.0.1:53083/s/"+strings.Repeat("a", 64) {
		t.Fatalf("create view must return the one-time URL: %q", got)
	}
}

func TestShareExpiryMustBeFutureAndBounded(t *testing.T) {
	if err := validateExpiry(time.Now().UTC().Add(-time.Minute)); err != ErrInvalidShareExpiry {
		t.Fatalf("past expiry = %v", err)
	}
	if err := validateExpiry(time.Now().UTC().Add(366 * 24 * time.Hour)); err != ErrInvalidShareExpiry {
		t.Fatalf("unbounded expiry = %v", err)
	}
}

type fakePasswordHasher struct{}

func (fakePasswordHasher) Make(password string) (string, error) {
	return "hash:" + password, nil
}

func (fakePasswordHasher) Check(password, hashed string) bool {
	return hashed == "hash:"+password
}

func TestSharePasswordIsOptionalButHasABoundedPolicy(t *testing.T) {
	hasher := fakePasswordHasher{}

	hash, err := makeSharePasswordHash(hasher, "")
	if err != nil || hash != "" {
		t.Fatalf("empty share password should remain unset: hash=%q err=%v", hash, err)
	}
	if _, err := makeSharePasswordHash(hasher, "short"); err != ErrInvalidSharePassword {
		t.Fatalf("short share password error = %v", err)
	}
	if _, err := makeSharePasswordHash(hasher, strings.Repeat("a", 73)); err != ErrInvalidSharePassword {
		t.Fatalf("long share password error = %v", err)
	}
	hash, err = makeSharePasswordHash(hasher, "correct horse")
	if err != nil || hash != "hash:correct horse" {
		t.Fatalf("share password should be hashed by the configured hasher: hash=%q err=%v", hash, err)
	}
}

func TestSharePasswordVerificationRejectsMissingAndWrongPasswords(t *testing.T) {
	hasher := fakePasswordHasher{}
	if err := verifySharePassword(hasher, "", "hash:correct horse"); err != ErrSharePasswordRequired {
		t.Fatalf("missing password error = %v", err)
	}
	if err := verifySharePassword(hasher, "wrong horse", "hash:correct horse"); err != ErrSharePasswordRequired {
		t.Fatalf("wrong password error = %v", err)
	}
	if err := verifySharePassword(hasher, "correct horse", "hash:correct horse"); err != nil {
		t.Fatalf("correct password error = %v", err)
	}
	if err := verifySharePassword(hasher, "", ""); err != nil {
		t.Fatalf("unprotected share error = %v", err)
	}
}
