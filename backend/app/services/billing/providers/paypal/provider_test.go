package paypal

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goravel/app/services/billing/providers"
)

func TestProviderCreatesAndQueriesCaptureOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/oauth2/token" {
			want := "Basic " + base64.StdEncoding.EncodeToString([]byte("client:secret"))
			if r.Header.Get("Authorization") != want {
				t.Fatal("missing OAuth basic auth")
			}
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":3600}`))
			return
		}
		if r.URL.Path == "/v2/checkout/orders" && r.Method == http.MethodPost {
			if r.Header.Get("PayPal-Request-Id") == "" || r.Header.Get("Authorization") != "Bearer token-1" {
				t.Fatal("order headers missing")
			}
			_, _ = w.Write([]byte(`{"id":"PP-ORDER-1","status":"CREATED","links":[{"rel":"approve","href":"https://paypal.test/approve/PP-ORDER-1"}]}`))
			return
		}
		if r.URL.Path == "/v2/checkout/orders/PP-ORDER-1" {
			_, _ = w.Write([]byte(`{"id":"PP-ORDER-1","status":"COMPLETED","purchase_units":[{"payments":{"captures":[{"id":"CAP-1","status":"COMPLETED","amount":{"currency_code":"USD","value":"19.99"}}]}}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	provider := New(Config{Enabled: true, BaseURL: server.URL, ClientID: "client", ClientSecret: "secret", WebhookID: "wh-1", Timeout: time.Second})
	session, err := provider.CreatePayment(context.Background(), providers.CreatePaymentRequest{OrderNo: "FST-1", PaymentIntentID: 7, AmountMinor: 1999, Currency: "USD", Description: "Creator", IdempotencyKey: "idem-1"})
	if err != nil || session.ProviderPaymentID != "PP-ORDER-1" || session.Status != "requires_action" || session.CheckoutURL == "" {
		t.Fatalf("session = %+v, err=%v", session, err)
	}
	payment, err := provider.QueryPayment(context.Background(), providers.QueryPaymentRequest{ProviderPaymentID: "PP-ORDER-1"})
	if err != nil || payment.Status != "succeeded" || payment.AmountMinor != 1999 || payment.ProviderPaymentTx != "CAP-1" {
		t.Fatalf("payment = %+v, err=%v", payment, err)
	}
}

func TestProviderVerifiesWebhookWithPayPalEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":3600}`))
			return
		}
		if r.URL.Path == "/v1/notifications/verify-webhook-signature" {
			_, _ = w.Write([]byte(`{"verification_status":"SUCCESS"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	provider := New(Config{Enabled: true, BaseURL: server.URL, ClientID: "client", ClientSecret: "secret", WebhookID: "wh-1", Timeout: time.Second})
	body := []byte(`{"id":"WH-1","event_type":"PAYMENT.CAPTURE.COMPLETED","resource":{"id":"CAP-1","amount":{"currency_code":"USD","value":"19.99"},"supplementary_data":{"related_ids":{"order_id":"PP-ORDER-1"}}}}`)
	event, err := provider.VerifyWebhook(context.Background(), providers.WebhookRequest{Headers: map[string]string{"paypal-transmission-id": "tx-1", "paypal-transmission-time": "2026-09-24T00:00:00Z", "paypal-cert-url": "https://api.test/cert", "paypal-auth-algo": "SHA256withRSA", "paypal-transmission-sig": "sig-1"}, Body: body})
	if err != nil || event.Status != "succeeded" || event.ProviderPaymentID != "PP-ORDER-1" || event.AmountMinor != 1999 {
		t.Fatalf("event = %+v, err=%v", event, err)
	}
}
