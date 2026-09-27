package billing

import "strings"

// GatewayURLSet contains the official provider defaults and site-derived
// callback routes. Persisted system settings may override these defaults.
type GatewayURLSet struct {
	BaseURL    string
	WebhookURL string
	ReturnURL  string
	CancelURL  string
}

func MissingGatewaySettings(provider string, values map[string]string) []string {
	required := map[string][]string{
		"paypal": {"payment.paypal.client_id", "payment.paypal.client_secret", "payment.paypal.webhook_id"},
		"xcash":  {"payment.xcash.app_id", "payment.xcash.hmac_key"},
		// The API key is required to create a NOWPayments invoice. The IPN
		// secret is only required by the webhook verifier and must not block
		// member checkout when webhook setup is still being completed.
		"nowpayments": {"payment.nowpayments.api_key"},
	}
	missing := make([]string, 0)
	for _, key := range required[strings.ToLower(strings.TrimSpace(provider))] {
		if strings.TrimSpace(values[key]) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func GatewayURLs(provider, environment, siteURL string) GatewayURLSet {
	site := strings.TrimRight(strings.TrimSpace(siteURL), "/")
	provider = strings.ToLower(strings.TrimSpace(provider))
	base := ""
	switch provider {
	case "paypal":
		if strings.EqualFold(strings.TrimSpace(environment), "production") || strings.EqualFold(strings.TrimSpace(environment), "live") {
			base = "https://api-m.paypal.com"
		} else {
			base = "https://api-m.sandbox.paypal.com"
		}
	case "xcash":
		base = "https://pay.xca.sh"
	case "nowpayments":
		base = "https://api.nowpayments.io"
	}
	return GatewayURLSet{
		BaseURL:    base,
		WebhookURL: site + "/api/v1/payment-gateways/" + provider + "/webhook",
		ReturnURL:  site + "/orders/{order_id}?payment=success",
		CancelURL:  site + "/orders/{order_id}?payment=cancelled",
	}
}
