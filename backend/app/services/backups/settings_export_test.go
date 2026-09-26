package backups

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportSettingsExcludesSecretsAndKeepsSafeValues(t *testing.T) {
	result := ExportSettings([]SettingValue{
		{Key: "site.title", Value: "FastImg"},
		{Key: "email.from_address", Value: "noreply@example.com"},
		{Key: "email.smtp.password", Value: "do-not-export"},
		{Key: "payment.nowpayments.api_key", Value: "api-secret"},
		{Key: "auth.turnstile.secret_key", Value: "turnstile-secret"},
		{Key: "custom.private_key", Value: "private-key"},
	})
	require.Equal(t, map[string]string{
		"site.title":         "FastImg",
		"email.from_address": "noreply@example.com",
	}, result.Values)
	require.ElementsMatch(t, []string{
		"email.smtp.password",
		"payment.nowpayments.api_key",
		"auth.turnstile.secret_key",
		"custom.private_key",
	}, result.ExcludedSecretKeys)
}
