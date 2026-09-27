package settingsservices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSetting(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		valueType string
		wantErr   bool
	}{
		{name: "string", key: "site.name", value: "Admin", valueType: "string"},
		{name: "secret value uses string validation", key: "payment.paypal.client_id", value: "", valueType: "secret"},
		{name: "boolean true", key: "feature.audit", value: "true", valueType: "boolean"},
		{name: "integer", key: "jobs.retry_limit", value: "3", valueType: "integer"},
		{name: "optional integer can be empty", key: "stats.retention_days", value: "", valueType: "integer"},
		{name: "json", key: "ui.theme", value: `{"mode":"dark"}`, valueType: "json"},
		{name: "invalid key", key: "site name", value: "Admin", valueType: "string", wantErr: true},
		{name: "invalid type", key: "site.name", value: "Admin", valueType: "url", wantErr: true},
		{name: "invalid boolean", key: "feature.audit", value: "yes", valueType: "boolean", wantErr: true},
		{name: "invalid integer", key: "jobs.retry_limit", value: "many", valueType: "integer", wantErr: true},
		{name: "invalid json", key: "ui.theme", value: "{", valueType: "json", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSetting(tt.key, tt.value, tt.valueType)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNormalizeSettingType(t *testing.T) {
	assert.Equal(t, "string", normalizeSettingType(""))
	assert.Equal(t, "boolean", normalizeSettingType(" boolean "))
}

func TestStorageProviderCredentialsAreSecretSettings(t *testing.T) {
	keys := []string{
		"storage.cloudflare_r2.access_key_id",
		"storage.cloudflare_r2.secret_access_key",
		"storage.aliyun_oss.access_key_id",
		"storage.aliyun_oss.access_key_secret",
		"storage.tencent_cos.secret_id",
		"storage.tencent_cos.secret_key",
	}
	for _, key := range keys {
		if !IsSecretKey(key) {
			t.Errorf("IsSecretKey(%q) = false, want true", key)
		}
	}
}
