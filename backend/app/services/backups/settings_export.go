package backups

import (
	"strings"

	settingsservices "goravel/app/services/settings"
)

type SettingValue struct {
	Key   string
	Value string
}

type SettingsExport struct {
	Values             map[string]string `json:"values"`
	ExcludedSecretKeys []string          `json:"excluded_secret_keys"`
}

func ExportSettings(settings []SettingValue) SettingsExport {
	result := SettingsExport{Values: make(map[string]string)}
	for _, setting := range settings {
		key := strings.TrimSpace(setting.Key)
		if key == "" {
			continue
		}
		if isBackupSecretKey(key) {
			result.ExcludedSecretKeys = append(result.ExcludedSecretKeys, key)
			continue
		}
		result.Values[key] = setting.Value
	}
	return result
}

func isBackupSecretKey(key string) bool {
	if settingsservices.IsSecretKey(key) {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(key))
	for _, marker := range []string{"password", "secret", "token", "private_key", "private-key", "api_key", "api-key", "access_key", "access-key", "client_secret", "client-secret"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
