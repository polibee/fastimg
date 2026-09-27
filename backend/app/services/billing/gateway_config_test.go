package billing

import "testing"

func TestGatewaySettingsRequireOnlyProviderCredentials(t *testing.T) {
	tests := []struct {
		provider string
		values   map[string]string
		want     []string
	}{
		{provider: "paypal", values: map[string]string{"payment.paypal.environment": "sandbox"}, want: []string{"payment.paypal.client_id", "payment.paypal.client_secret", "payment.paypal.webhook_id"}},
		{provider: "xcash", values: map[string]string{}, want: []string{"payment.xcash.app_id", "payment.xcash.hmac_key"}},
		{provider: "nowpayments", values: map[string]string{"payment.nowpayments.api_key": "key"}, want: []string{}},
	}
	for _, tt := range tests {
		got := MissingGatewaySettings(tt.provider, tt.values)
		if len(got) != len(tt.want) {
			t.Fatalf("%s missing settings = %#v, want %#v", tt.provider, got, tt.want)
		}
		for index := range tt.want {
			if got[index] != tt.want[index] {
				t.Fatalf("%s missing settings = %#v, want %#v", tt.provider, got, tt.want)
			}
		}
	}
}

func TestGatewayURLsUseFixedProviderEndpointsAndSiteRoutes(t *testing.T) {
	paypal := GatewayURLs("paypal", "sandbox", "https://img.example.com")
	if paypal.BaseURL != "https://api-m.sandbox.paypal.com" || paypal.WebhookURL != "https://img.example.com/api/v1/payment-gateways/paypal/webhook" {
		t.Fatalf("paypal urls = %#v", paypal)
	}
	xcash := GatewayURLs("xcash", "", "https://img.example.com/")
	if xcash.BaseURL != "https://pay.xca.sh" || xcash.ReturnURL != "https://img.example.com/orders/{order_id}?payment=success" {
		t.Fatalf("xcash urls = %#v", xcash)
	}
	nowpayments := GatewayURLs("nowpayments", "", "https://img.example.com")
	if nowpayments.BaseURL != "https://api.nowpayments.io" || nowpayments.CancelURL != "https://img.example.com/orders/{order_id}?payment=cancelled" {
		t.Fatalf("nowpayments urls = %#v", nowpayments)
	}
}
