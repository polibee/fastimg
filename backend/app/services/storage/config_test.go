package storage

import (
	"testing"
)

func TestNormalizeConnectionConfigDerivesR2Endpoint(t *testing.T) {
	config, err := NormalizeConnectionConfig(ConnectionConfig{
		ProviderCode:    ProviderCloudflareR2,
		AccountID:       "abc123",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Bucket:          "media",
	})
	if err != nil {
		t.Fatalf("NormalizeConnectionConfig() error = %v", err)
	}
	if config.Region != "auto" {
		t.Fatalf("R2 region = %q, want auto", config.Region)
	}
	if config.Endpoint != "https://abc123.r2.cloudflarestorage.com" {
		t.Fatalf("R2 endpoint = %q", config.Endpoint)
	}
}

func TestNormalizeConnectionConfigDerivesJurisdictionR2Endpoint(t *testing.T) {
	config, err := NormalizeConnectionConfig(ConnectionConfig{
		ProviderCode:    ProviderCloudflareR2,
		AccountID:       "abc123",
		Jurisdiction:    "eu",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Bucket:          "media",
	})
	if err != nil {
		t.Fatalf("NormalizeConnectionConfig() error = %v", err)
	}
	if config.Endpoint != "https://abc123.eu.r2.cloudflarestorage.com" {
		t.Fatalf("jurisdiction R2 endpoint = %q", config.Endpoint)
	}
}

func TestNormalizeConnectionConfigDerivesAliyunOSSEndpoint(t *testing.T) {
	config, err := NormalizeConnectionConfig(ConnectionConfig{
		ProviderCode:    ProviderAliyunOSS,
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Bucket:          "media",
		Region:          "cn-hangzhou",
	})
	if err != nil {
		t.Fatalf("NormalizeConnectionConfig() error = %v", err)
	}
	if config.Endpoint != "https://oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("OSS endpoint = %q", config.Endpoint)
	}
	if config.PublicBaseURL != "https://media.oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("OSS public URL = %q", config.PublicBaseURL)
	}
}

func TestNormalizeConnectionConfigDerivesTencentCOSBucketAndEndpoint(t *testing.T) {
	config, err := NormalizeConnectionConfig(ConnectionConfig{
		ProviderCode: ProviderTencentCOS,
		AppID:        "1250000000",
		SecretID:     "secret-id",
		SecretKey:    "secret-key",
		Bucket:       "media",
		Region:       "ap-guangzhou",
	})
	if err != nil {
		t.Fatalf("NormalizeConnectionConfig() error = %v", err)
	}
	if config.Bucket != "media-1250000000" {
		t.Fatalf("COS bucket = %q", config.Bucket)
	}
	if config.Endpoint != "https://media-1250000000.cos.ap-guangzhou.myqcloud.com" {
		t.Fatalf("COS endpoint = %q", config.Endpoint)
	}
	if config.PublicBaseURL != config.Endpoint {
		t.Fatalf("COS public URL = %q, want endpoint %q", config.PublicBaseURL, config.Endpoint)
	}
}

func TestNormalizeConnectionConfigRejectsMissingProviderCredentials(t *testing.T) {
	tests := []struct {
		name   string
		config ConnectionConfig
		want   string
	}{
		{
			name:   "unknown provider",
			config: ConnectionConfig{ProviderCode: "unknown"},
			want:   "unsupported provider",
		},
		{
			name:   "r2 missing account",
			config: ConnectionConfig{ProviderCode: ProviderCloudflareR2, AccessKeyID: "a", SecretAccessKey: "s", Bucket: "b"},
			want:   "account_id is required",
		},
		{
			name:   "oss missing region",
			config: ConnectionConfig{ProviderCode: ProviderAliyunOSS, AccessKeyID: "a", SecretAccessKey: "s", Bucket: "b"},
			want:   "region is required",
		},
		{
			name:   "cos missing app id",
			config: ConnectionConfig{ProviderCode: ProviderTencentCOS, SecretID: "id", SecretKey: "key", Bucket: "b", Region: "ap-guangzhou"},
			want:   "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeConnectionConfig(tt.config)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestNormalizeConnectionConfigPreservesCustomPublicBaseURL(t *testing.T) {
	config, err := NormalizeConnectionConfig(ConnectionConfig{
		ProviderCode:    ProviderCloudflareR2,
		AccountID:       "abc123",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Bucket:          "media",
		PublicBaseURL:   "https://img.example.com/",
	})
	if err != nil {
		t.Fatalf("NormalizeConnectionConfig() error = %v", err)
	}
	if config.PublicBaseURL != "https://img.example.com" {
		t.Fatalf("custom public URL = %q", config.PublicBaseURL)
	}
}
