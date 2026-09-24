package xcash

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"goravel/app/services/billing/providers"
)

func TestProviderUsesSignedCreateAndPublicQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/invoice" && r.Method == http.MethodPost {
			if r.Header.Get("XC-Appid") != "app-1" || r.Header.Get("XC-Signature") == "" {
				t.Fatalf("signed headers missing: %+v", r.Header)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sys_no":"INV1","out_no":"FST-1","pay_url":"https://pay.test/INV1","status":"waiting"}`))
			return
		}
		if r.URL.Path == "/v1/invoice/INV1" && r.Method == http.MethodGet {
			if r.Header.Get("XC-Signature") != "" {
				t.Fatal("public query must not include a signature")
			}
			_, _ = w.Write([]byte(`{"sys_no":"INV1","amount":"19.99","currency":"CNY","status":"completed","payment":{"hash":"0xabc"}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	provider := New(Config{Enabled: true, BaseURL: server.URL, AppID: "app-1", HMACKey: "key-1", Timeout: time.Second})
	session, err := provider.CreatePayment(context.Background(), providers.CreatePaymentRequest{OrderNo: "FST-1", PaymentIntentID: 7, AmountMinor: 1999, Currency: "CNY", Description: "Creator"})
	if err != nil || session.ProviderPaymentID != "INV1" || session.Status != "pending" {
		t.Fatalf("create session = %+v, err=%v", session, err)
	}
	payment, err := provider.QueryPayment(context.Background(), providers.QueryPaymentRequest{ProviderPaymentID: "INV1"})
	if err != nil || payment.Status != "succeeded" || payment.AmountMinor != 1999 || payment.ProviderPaymentTx != "0xabc" {
		t.Fatalf("query payment = %+v, err=%v", payment, err)
	}
}

func TestProviderWebhookQueriesFiatAmountBeforeReturningEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoice/INV1" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"sys_no":"INV1","amount":"19.99","currency":"CNY","status":"completed","payment":{"hash":"0xabc"}}`))
	}))
	defer server.Close()

	provider := New(Config{Enabled: true, BaseURL: server.URL, AppID: "app-1", HMACKey: "key-1", Timeout: time.Second})
	body := []byte(`{"type":"invoice","data":{"sys_no":"INV1","hash":"0xabc","confirmed":true}}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	event, err := provider.VerifyWebhook(context.Background(), providers.WebhookRequest{Headers: map[string]string{"XC-Appid": "app-1", "XC-Timestamp": timestamp, "XC-Nonce": "event-1", "XC-Signature": Sign("key-1", "event-1", timestamp, body)}, Body: body})
	if err != nil || event.Status != "succeeded" || event.AmountMinor != 1999 || event.Currency != "CNY" {
		t.Fatalf("webhook event = %+v, err=%v", event, err)
	}
}

func TestDisabledProviderNeverCallsNetwork(t *testing.T) {
	provider := New(Config{BaseURL: "http://127.0.0.1:1", AppID: "app", HMACKey: "key"})
	_, err := provider.CreatePayment(context.Background(), providers.CreatePaymentRequest{OrderNo: "FST-1", PaymentIntentID: 1, AmountMinor: 100, Currency: "CNY"})
	if err != providers.ErrGatewayUnavailable {
		t.Fatalf("disabled provider error = %v, want gateway unavailable", err)
	}
}
