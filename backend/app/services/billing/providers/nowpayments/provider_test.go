package nowpayments

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goravel/app/services/billing/providers"
)

func TestProviderCreatesInvoiceAndQueriesPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "api-key" {
			t.Fatal("missing API key")
		}
		if r.URL.Path == "/v1/invoice" {
			_, _ = w.Write([]byte(`{"id":12345,"invoice_url":"https://now.test/i/12345","payment_status":"waiting"}`))
			return
		}
		if r.URL.Path == "/v1/payment/12345" {
			_, _ = w.Write([]byte(`{"payment_id":12345,"payment_status":"finished","price_amount":19.99,"price_currency":"cny","pay_currency":"usdttrc20","actually_paid":19.99,"pay_address":"T1"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	provider := New(Config{Enabled: true, BaseURL: server.URL, APIKey: "api-key", IPNSecret: "ipn", Timeout: time.Second})
	session, err := provider.CreatePayment(context.Background(), providers.CreatePaymentRequest{OrderNo: "FST-1", PaymentIntentID: 7, AmountMinor: 1999, Currency: "CNY", Description: "Creator"})
	if err != nil || session.ProviderPaymentID != "12345" || session.Status != "pending" || session.CheckoutURL == "" {
		t.Fatalf("create session = %+v, err=%v", session, err)
	}
	payment, err := provider.QueryPayment(context.Background(), providers.QueryPaymentRequest{ProviderPaymentID: "12345"})
	if err != nil || payment.Status != "succeeded" || payment.AmountMinor != 1999 || payment.Currency != "CNY" {
		t.Fatalf("query payment = %+v, err=%v", payment, err)
	}
}

func TestVerifyIPNUsesCanonicalJSONAndRejectsPartialPayment(t *testing.T) {
	provider := New(Config{Enabled: true, BaseURL: "http://127.0.0.1:1", APIKey: "api-key", IPNSecret: "ipn"})
	body := []byte(`{"price_currency":"CNY","price_amount":19.99,"payment_status":"partially_paid","payment_id":12345,"order_id":"FST-1"}`)
	mac := hmac.New(sha512.New, []byte("ipn"))
	_, _ = mac.Write([]byte(`{"order_id":"FST-1","payment_id":12345,"payment_status":"partially_paid","price_amount":19.99,"price_currency":"CNY"}`))
	event, err := provider.VerifyWebhook(context.Background(), providers.WebhookRequest{Headers: map[string]string{"x-nowpayments-sig": hex.EncodeToString(mac.Sum(nil))}, Body: body})
	if err != nil || event.Status != "pending_review" || event.ProviderPaymentID != "12345" {
		t.Fatalf("event = %+v, err=%v", event, err)
	}
}
