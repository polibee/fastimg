package billing

import (
	"sync"
	"time"

	"goravel/app/facades"
	"goravel/app/services/billing/providers"
	"goravel/app/services/billing/providers/fake"
	"goravel/app/services/billing/providers/nowpayments"
	"goravel/app/services/billing/providers/paypal"
	"goravel/app/services/billing/providers/xcash"
)

var (
	developmentRegistryOnce sync.Once
	developmentRegistry     *providers.Registry
)

// DefaultGatewayRegistry is process-scoped so member checkout and webhook
// handling use the same Fake Provider instance during local development.
func DefaultGatewayRegistry() *providers.Registry {
	developmentRegistryOnce.Do(func() {
		developmentRegistry = providers.NewRegistry()
		_ = developmentRegistry.Register("fake", fake.New())
		if facades.Config().GetString("payment.providers.xcash.enabled", "false") == "true" {
			timeout, err := time.ParseDuration(facades.Config().GetString("payment.providers.xcash.timeout", "10s"))
			if err != nil {
				timeout = 10 * time.Second
			}
			_ = developmentRegistry.Register("xcash", xcash.New(xcash.Config{
				Enabled: true, BaseURL: facades.Config().GetString("payment.providers.xcash.base_url", ""),
				AppID: facades.Config().GetString("payment.providers.xcash.app_id", ""), HMACKey: facades.Config().GetString("payment.providers.xcash.hmac_key", ""),
				Timeout: timeout, CallbackURL: facades.Config().GetString("payment.providers.xcash.callback_url", ""), ReturnURL: facades.Config().GetString("payment.providers.xcash.return_url", ""),
			}))
		}
		if facades.Config().GetString("payment.providers.nowpayments.enabled", "false") == "true" {
			timeout, err := time.ParseDuration(facades.Config().GetString("payment.providers.nowpayments.timeout", "10s"))
			if err != nil {
				timeout = 10 * time.Second
			}
			_ = developmentRegistry.Register("nowpayments", nowpayments.New(nowpayments.Config{
				Enabled: true, BaseURL: facades.Config().GetString("payment.providers.nowpayments.base_url", ""),
				APIKey: facades.Config().GetString("payment.providers.nowpayments.api_key", ""), IPNSecret: facades.Config().GetString("payment.providers.nowpayments.ipn_secret", ""),
				Timeout: timeout, CallbackURL: facades.Config().GetString("payment.providers.nowpayments.callback_url", ""),
				SuccessURL: facades.Config().GetString("payment.providers.nowpayments.success_url", ""), CancelURL: facades.Config().GetString("payment.providers.nowpayments.cancel_url", ""),
			}))
		}
		if facades.Config().GetString("payment.providers.paypal.enabled", "false") == "true" {
			timeout, err := time.ParseDuration(facades.Config().GetString("payment.providers.paypal.timeout", "10s"))
			if err != nil {
				timeout = 10 * time.Second
			}
			_ = developmentRegistry.Register("paypal", paypal.New(paypal.Config{
				Enabled: true, Environment: facades.Config().GetString("payment.providers.paypal.environment", "sandbox"), BaseURL: facades.Config().GetString("payment.providers.paypal.base_url", ""),
				ClientID: facades.Config().GetString("payment.providers.paypal.client_id", ""), ClientSecret: facades.Config().GetString("payment.providers.paypal.client_secret", ""), WebhookID: facades.Config().GetString("payment.providers.paypal.webhook_id", ""),
				Timeout: timeout, ReturnURL: facades.Config().GetString("payment.providers.paypal.return_url", ""), CancelURL: facades.Config().GetString("payment.providers.paypal.cancel_url", ""),
			}))
		}
	})
	return developmentRegistry
}
