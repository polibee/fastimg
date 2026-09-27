package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("cors", map[string]any{
		// Cross-Origin Resource Sharing (CORS) Configuration
		//
		// Here you may configure your settings for cross-origin resource sharing
		// or "CORS". This determines what cross-origin operations may execute
		// in web browsers. You are free to adjust these settings as needed.
		//
		// To learn more: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
		"paths":                []string{},
		// The FastImg middleware applies the exact CORS_ALLOWED_ORIGINS
		// allowlist. Keep the framework fallback fail-closed as well; a wildcard
		// origin is unsafe for authenticated browser requests.
		"allowed_methods":      []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		"allowed_origins":      []string{},
		"allowed_headers":      []string{"Content-Type", "Authorization", "X-API-Key", "Idempotency-Key"},
		"exposed_headers":      []string{},
		"max_age":              0,
		"supports_credentials": false,
	})
}
