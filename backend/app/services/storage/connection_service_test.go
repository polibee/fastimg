package storage

import "testing"

func TestProviderDefinitionsExposeRequiredConfiguration(t *testing.T) {
	definitions := ProviderDefinitions()
	if len(definitions) != 4 {
		t.Fatalf("expected local and three object storage providers, got %d", len(definitions))
	}
	r2 := definitions[ProviderCloudflareR2]
	if r2 == nil || !hasConfigField(r2, "access_key_id", true) || !hasConfigField(r2, "secret_access_key", true) || !hasConfigField(r2, "account_id", false) {
		t.Fatalf("R2 definition does not declare required credential fields: %+v", r2)
	}
	if !hasConfigField(definitions[ProviderAliyunOSS], "access_key_secret", true) {
		t.Fatal("OSS secret field must be marked secret")
	}
	if !hasConfigField(definitions[ProviderTencentCOS], "app_id", false) {
		t.Fatal("COS AppID must be a visible configuration field")
	}
}

func hasConfigField(definition *ProviderDefinition, name string, secret bool) bool {
	if definition == nil {
		return false
	}
	for _, field := range definition.Fields {
		if field.Name == name {
			return field.Secret == secret
		}
	}
	return false
}

func TestConnectionConfigMapPreservesConfiguredSecrets(t *testing.T) {
	existing := ConnectionConfig{
		ProviderCode:    ProviderCloudflareR2,
		AccountID:       "account",
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-value",
		Bucket:          "images",
		PublicBaseURL:   "https://img.example.com",
	}
	merged, err := configFromValues(existing, map[string]string{
		"account_id":        "account",
		"access_key_id":     "__configured__",
		"secret_access_key": "",
		"bucket":            "images-v2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if merged.AccessKeyID != existing.AccessKeyID || merged.SecretAccessKey != existing.SecretAccessKey {
		t.Fatalf("configured secrets were not preserved: %+v", merged)
	}
	if merged.Bucket != "images-v2" {
		t.Fatalf("visible values were not updated: %+v", merged)
	}
}

func TestConnectionStatusRequiresCloudConnectionTest(t *testing.T) {
	status, code := statusForProvider(ProviderCloudflareR2, true, nil)
	if status != "degraded" || code != "STORAGE_CONNECTION_UNTESTED" {
		t.Fatalf("status = %q/%q, want degraded/untested", status, code)
	}
	status, code = statusForProvider(ProviderLocal, true, nil)
	if status != "healthy" || code != "" {
		t.Fatalf("local status = %q/%q, want healthy", status, code)
	}
}

func TestCloudProviderCanBecomePrimaryWhenAdapterIsAvailable(t *testing.T) {
	if err := validatePrimarySelection(ProviderCloudflareR2, true); err != nil {
		t.Fatalf("primary selection error = %v, want nil", err)
	}
	if err := validatePrimarySelection(ProviderLocal, true); err != nil {
		t.Fatalf("local primary selection failed: %v", err)
	}
}
