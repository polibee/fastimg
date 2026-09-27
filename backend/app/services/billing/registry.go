package billing

import (
	"strings"
	"sync"
	"time"

	"goravel/app/facades"
	"goravel/app/services/billing/providers"
	"goravel/app/services/billing/providers/fake"
	"goravel/app/services/billing/providers/nowpayments"
	"goravel/app/services/billing/providers/paypal"
	"goravel/app/services/billing/providers/xcash"
	settingsservices "goravel/app/services/settings"
)

var (
	developmentRegistryMu sync.Mutex
	developmentRegistry   *providers.Registry
	developmentFake       *fake.Provider
)

// DefaultGatewayRegistry is process-scoped so member checkout and webhook
// handling use the same registry and Fake Provider instance during local
// development. Provider settings are refreshed on each request by the
// controllers; saving settings no longer requires a backend restart.
func DefaultGatewayRegistry() *providers.Registry {
	developmentRegistryMu.Lock()
	if developmentRegistry == nil {
		developmentRegistry = providers.NewRegistry()
		developmentFake = fake.New()
	}
	developmentRegistryMu.Unlock()
	RefreshGatewayRegistry()
	return developmentRegistry
}

// RefreshGatewayRegistry reloads enabled providers from system settings while
// preserving the process-local fake provider used by local payment tests.
func RefreshGatewayRegistry() {
	developmentRegistryMu.Lock()
	defer developmentRegistryMu.Unlock()
	if developmentRegistry == nil {
		developmentRegistry = providers.NewRegistry()
		developmentFake = fake.New()
	}

	configured := make(map[string]providers.PaymentGateway)
	if settingOrDefault("payment.fake.enabled", fakeGatewayDefaultEnabled(facades.Config().GetString("app.env", "production"))) {
		configured["fake"] = developmentFake
	}
	if settingOrDefault("payment.xcash.enabled", facades.Config().GetString("payment.providers.xcash.enabled", "false") == "true") {
		appID := settingOrConfig("payment.xcash.app_id", "payment.providers.xcash.app_id", "")
		hmacKey := settingOrConfig("payment.xcash.hmac_key", "payment.providers.xcash.hmac_key", "")
		siteURL := settingOrConfig("site_url", "app.url", facades.Config().GetString("app.url", ""))
		urls := GatewayURLs("xcash", "", siteURL)
		urls.BaseURL = settingOrConfig("payment.xcash.base_url", "payment.providers.xcash.base_url", urls.BaseURL)
		urls.WebhookURL = settingOrConfig("payment.xcash.callback_url", "payment.providers.xcash.callback_url", urls.WebhookURL)
		urls.ReturnURL = settingOrConfig("payment.xcash.return_url", "payment.providers.xcash.return_url", urls.ReturnURL)
		if strings.TrimSpace(appID) != "" && strings.TrimSpace(hmacKey) != "" {
			configured["xcash"] = xcash.New(xcash.Config{Enabled: true, BaseURL: urls.BaseURL, AppID: appID, HMACKey: hmacKey, Timeout: providerTimeout("payment.providers.xcash.timeout"), CallbackURL: urls.WebhookURL, ReturnURL: urls.ReturnURL})
		}
	}
	if settingOrDefault("payment.nowpayments.enabled", facades.Config().GetString("payment.providers.nowpayments.enabled", "false") == "true") {
		apiKey := settingOrConfig("payment.nowpayments.api_key", "payment.providers.nowpayments.api_key", "")
		ipnSecret := settingOrConfig("payment.nowpayments.ipn_secret", "payment.providers.nowpayments.ipn_secret", "")
		siteURL := settingOrConfig("site_url", "app.url", facades.Config().GetString("app.url", ""))
		urls := GatewayURLs("nowpayments", "", siteURL)
		urls.BaseURL = settingOrConfig("payment.nowpayments.base_url", "payment.providers.nowpayments.base_url", urls.BaseURL)
		urls.WebhookURL = settingOrConfig("payment.nowpayments.callback_url", "payment.providers.nowpayments.callback_url", urls.WebhookURL)
		urls.ReturnURL = settingOrConfig("payment.nowpayments.success_url", "payment.providers.nowpayments.success_url", urls.ReturnURL)
		urls.CancelURL = settingOrConfig("payment.nowpayments.cancel_url", "payment.providers.nowpayments.cancel_url", urls.CancelURL)
		// NOWPayments uses API Key for outbound invoice/status calls. IPN Secret
		// is only needed for inbound callback verification, so it must not hide
		// an otherwise usable checkout channel.
		if strings.TrimSpace(apiKey) != "" {
			configured["nowpayments"] = nowpayments.New(nowpayments.Config{Enabled: true, BaseURL: urls.BaseURL, APIKey: apiKey, IPNSecret: ipnSecret, Timeout: providerTimeout("payment.providers.nowpayments.timeout"), CallbackURL: urls.WebhookURL, SuccessURL: urls.ReturnURL, CancelURL: urls.CancelURL})
		}
	}
	if settingOrDefault("payment.paypal.enabled", facades.Config().GetString("payment.providers.paypal.enabled", "false") == "true") {
		clientID := settingOrConfig("payment.paypal.client_id", "payment.providers.paypal.client_id", "")
		clientSecret := settingOrConfig("payment.paypal.client_secret", "payment.providers.paypal.client_secret", "")
		webhookID := settingOrConfig("payment.paypal.webhook_id", "payment.providers.paypal.webhook_id", "")
		environment := settingOrConfig("payment.paypal.environment", "payment.providers.paypal.environment", "sandbox")
		siteURL := settingOrConfig("site_url", "app.url", facades.Config().GetString("app.url", ""))
		urls := GatewayURLs("paypal", environment, siteURL)
		urls.BaseURL = settingOrConfig("payment.paypal.base_url", "payment.providers.paypal.base_url", urls.BaseURL)
		urls.ReturnURL = settingOrConfig("payment.paypal.return_url", "payment.providers.paypal.return_url", urls.ReturnURL)
		urls.CancelURL = settingOrConfig("payment.paypal.cancel_url", "payment.providers.paypal.cancel_url", urls.CancelURL)
		if strings.TrimSpace(clientID) != "" && strings.TrimSpace(clientSecret) != "" && strings.TrimSpace(webhookID) != "" {
			configured["paypal"] = paypal.New(paypal.Config{Enabled: true, Environment: environment, BaseURL: urls.BaseURL, ClientID: clientID, ClientSecret: clientSecret, WebhookID: webhookID, Timeout: providerTimeout("payment.providers.paypal.timeout"), ReturnURL: urls.ReturnURL, CancelURL: urls.CancelURL})
		}
	}
	developmentRegistry.Replace(configured)
}

func fakeGatewayDefaultEnabled(environment string) bool {
	return strings.ToLower(strings.TrimSpace(environment)) != "production"
}

func providerTimeout(key string) time.Duration {
	timeout, err := time.ParseDuration(facades.Config().GetString(key, "10s"))
	if err != nil || timeout <= 0 {
		return 10 * time.Second
	}
	return timeout
}

func settingOrDefault(key string, fallback bool) bool {
	if !facades.Schema().HasTable("system_settings") {
		return fallback
	}
	var setting struct {
		Value string `orm:"value"`
	}
	if err := facades.Orm().Query().Table("system_settings").Where("key = ?", key).First(&setting); err != nil || strings.TrimSpace(setting.Value) == "" {
		return fallback
	}
	return strings.EqualFold(strings.TrimSpace(setting.Value), "true")
}

func settingOrConfig(settingKey, configKey, fallback string) string {
	if value := settingsservices.NewSettingService().Resolve(settingKey, ""); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return facades.Config().GetString(configKey, fallback)
}
