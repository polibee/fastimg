package xcash

import (
	"strings"
	"testing"
	"time"
)

func TestSignUsesNonceTimestampAndExactBody(t *testing.T) {
	got := Sign("secret", "nonce-1", "1700000000", []byte(`{"out_no":"FST-1"}`))
	const want = "441bff974e727aa3d299d0e8afe1defd3eb38dbdfc5e92299911ee77402c67ee"
	if got != want {
		t.Fatalf("signature = %q, want %q", got, want)
	}
	if len(got) != 64 || strings.ToLower(got) != got {
		t.Fatalf("signature = %q, want lowercase hex sha256", got)
	}
	if !VerifySignature("secret", map[string]string{
		"XC-Timestamp": "1700000000", "XC-Nonce": "nonce-1", "XC-Signature": got,
	}, []byte(`{"out_no":"FST-1"}`), time.Unix(1700000000, 0), 300*time.Second) {
		t.Fatal("signature did not verify")
	}
}

func TestVerifySignatureRejectsMissingOrStaleHeaders(t *testing.T) {
	now := time.Unix(1700000000, 0)
	body := []byte(`{"ok":true}`)
	signature := Sign("secret", "nonce-1", "1700000000", body)
	if VerifySignature("secret", map[string]string{"XC-Timestamp": "1700000000", "XC-Signature": signature}, body, now, 300*time.Second) {
		t.Fatal("missing nonce must be rejected")
	}
	if VerifySignature("secret", map[string]string{"XC-Timestamp": "1699999000", "XC-Nonce": "nonce-1", "XC-Signature": signature}, body, now, 300*time.Second) {
		t.Fatal("stale timestamp must be rejected")
	}
}

func TestMapStatusAndParseAmount(t *testing.T) {
	for _, item := range []struct{ input, want string }{
		{"waiting", "pending"}, {"completed", "succeeded"}, {"expired", "failed"}, {"underpaid", "pending_review"}, {"overpaid", "pending_review"}, {"wrong_network", "failed"},
	} {
		if got := MapStatus(item.input, false, ""); got != item.want {
			t.Fatalf("MapStatus(%q) = %q, want %q", item.input, got, item.want)
		}
	}
	if got := MapStatus("waiting", true, "low"); got != "succeeded" {
		t.Fatalf("confirmed status = %q, want succeeded", got)
	}
	if got, err := ParseMinor("29.99"); err != nil || got != 2999 {
		t.Fatalf("ParseMinor(29.99) = %d, %v", got, err)
	}
}
