package emailservices

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAliyunSignatureIsStableAndExcludesSignatureField(t *testing.T) {
	params := map[string]string{
		"Action":          "SingleSendMail",
		"AccessKeyId":     "test-key",
		"AccountName":     "noreply@example.com",
		"AddressType":     "1",
		"Format":          "JSON",
		"SignatureMethod": "HMAC-SHA1",
		"SignatureNonce":  "nonce",
		"SignatureVersion": "1.0",
		"Timestamp":       "2026-09-26T00:00:00Z",
		"Version":         "2015-11-23",
	}

	first := aliyunSignature("POST", params, "test-secret")
	params["Signature"] = "must-not-be-signed"
	second := aliyunSignature("POST", params, "test-secret")
	if first == "" || first != second {
		t.Fatalf("signature changed when Signature parameter was added: %q != %q", first, second)
	}
}

func TestValidMailboxRejectsHeaderInjection(t *testing.T) {
	for _, value := range []string{"user@example.com\r\nBcc: attacker@example.com", "Display <user@example.com>", "not-an-email"} {
		if validMailbox(value) {
			t.Fatalf("expected mailbox to be rejected: %q", value)
		}
	}
	if !validMailbox("user@example.com") {
		t.Fatal("expected a plain mailbox to be accepted")
	}
}

func TestFormatAddress(t *testing.T) {
	if got := formatAddress("noreply@example.com", "FastImg"); got != "FastImg <noreply@example.com>" {
		t.Fatalf("unexpected formatted address: %q", got)
	}
	if got := formatAddress("noreply@example.com", ""); got != "noreply@example.com" {
		t.Fatalf("unexpected address without name: %q", got)
	}
}

func TestResendProviderUsesOfficialEmailPayload(t *testing.T) {
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer re_test" {
			t.Fatalf("unexpected Resend request: %s %s", request.Method, request.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			t.Fatal(err)
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{"id":"email-id"}`))
	}))
	defer server.Close()

	values := map[string]string{
		"email.enabled": "true", "email.provider": ProviderResend,
		"email.from_address": "noreply@example.com", "email.from_name": "FastImg",
		"email.resend.api_key": "re_test", "email.resend.endpoint": server.URL,
	}
	service := &Service{resolve: func(key, fallback string) string { if value, ok := values[key]; ok { return value }; return fallback }, client: server.Client()}
	if err := service.Send(context.Background(), Message{To: "user@example.com", Subject: "Verify", Text: "text", HTML: "<p>html</p>"}); err != nil {
		t.Fatal(err)
	}
	if requestBody["from"] != "FastImg <noreply@example.com>" || requestBody["subject"] != "Verify" {
		t.Fatalf("unexpected Resend payload: %#v", requestBody)
	}
}

func TestAliyunProviderPostsSignedSingleSendMailRequest(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		form, _ = url.ParseQuery(string(body))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"RequestId":"request-id"}`))
	}))
	defer server.Close()

	values := map[string]string{
		"email.enabled": "true", "email.provider": ProviderAliyun,
		"email.from_address": "noreply@example.com", "email.from_name": "FastImg",
		"email.aliyun.access_key_id": "access-id", "email.aliyun.access_key_secret": "access-secret",
		"email.aliyun.account_name": "noreply@example.com", "email.aliyun.endpoint": server.URL,
	}
	service := &Service{resolve: func(key, fallback string) string { if value, ok := values[key]; ok { return value }; return fallback }, client: server.Client(), now: func() time.Time { return time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC) }}
	if err := service.Send(context.Background(), Message{To: "user@example.com", Subject: "Verify", Text: "text", HTML: "<p>html</p>"}); err != nil {
		t.Fatal(err)
	}
	if form.Get("Action") != "SingleSendMail" || form.Get("Version") != "2015-11-23" || form.Get("Signature") == "" || !strings.Contains(form.Get("ToAddress"), "user@example.com") {
		t.Fatalf("unexpected Aliyun request: %#v", form)
	}
}
