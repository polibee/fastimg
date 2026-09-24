package config

import "goravel/app/facades"

func init() {
	config := facades.Config()
	config.Add("payment", map[string]any{
		"providers": map[string]any{
			"xcash": map[string]any{
				"enabled":      config.Env("XCASH_ENABLED", "false"),
				"base_url":     config.Env("XCASH_API_BASE_URL", "https://pay.xca.sh"),
				"app_id":       config.Env("XCASH_APP_ID", ""),
				"hmac_key":     config.Env("XCASH_HMAC_KEY", ""),
				"timeout":      config.Env("XCASH_TIMEOUT", "10s"),
				"callback_url": config.Env("XCASH_CALLBACK_URL", ""),
				"return_url":   config.Env("XCASH_RETURN_URL", ""),
			},
			"nowpayments": map[string]any{
				"enabled":      config.Env("NOWPAYMENTS_ENABLED", "false"),
				"base_url":     config.Env("NOWPAYMENTS_API_BASE_URL", "https://api.nowpayments.io"),
				"api_key":      config.Env("NOWPAYMENTS_API_KEY", ""),
				"ipn_secret":   config.Env("NOWPAYMENTS_IPN_SECRET", ""),
				"timeout":      config.Env("NOWPAYMENTS_TIMEOUT", "10s"),
				"callback_url": config.Env("NOWPAYMENTS_CALLBACK_URL", ""),
				"success_url":  config.Env("NOWPAYMENTS_SUCCESS_URL", ""),
				"cancel_url":   config.Env("NOWPAYMENTS_CANCEL_URL", ""),
			},
			"paypal": map[string]any{
				"enabled":       config.Env("PAYPAL_ENABLED", "false"),
				"environment":   config.Env("PAYPAL_ENVIRONMENT", "sandbox"),
				"base_url":      config.Env("PAYPAL_API_BASE_URL", "https://api-m.sandbox.paypal.com"),
				"client_id":     config.Env("PAYPAL_CLIENT_ID", ""),
				"client_secret": config.Env("PAYPAL_CLIENT_SECRET", ""),
				"webhook_id":    config.Env("PAYPAL_WEBHOOK_ID", ""),
				"timeout":       config.Env("PAYPAL_TIMEOUT", "10s"),
				"return_url":    config.Env("PAYPAL_RETURN_URL", ""),
				"cancel_url":    config.Env("PAYPAL_CANCEL_URL", ""),
			},
		},
	})
}
