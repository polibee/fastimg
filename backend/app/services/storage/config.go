package storage

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	ProviderLocal        = "local"
	ProviderCloudflareR2 = "cloudflare_r2"
	ProviderAliyunOSS    = "aliyun_oss"
	ProviderTencentCOS   = "tencent_cos"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrInvalidEndpoint     = errors.New("endpoint must be an absolute http or https URL")
)

// ConnectionConfig is the provider-neutral configuration boundary. Provider
// adapters may use provider-specific fields, but callers do not need to know
// how endpoint and bucket URLs are derived.
type ConnectionConfig struct {
	ProviderCode string
	AccountID    string
	Jurisdiction string

	AccessKeyID     string
	SecretAccessKey string

	Bucket    string
	Region    string
	AppID     string
	SecretID  string
	SecretKey string

	Endpoint      string
	PublicBaseURL string
}

// NormalizeConnectionConfig validates the minimum credentials for a provider
// and fills provider-specific defaults used by runtime adapters. It does not
// contact a remote service.
func NormalizeConnectionConfig(input ConnectionConfig) (ConnectionConfig, error) {
	config := input
	config.ProviderCode = strings.ToLower(strings.TrimSpace(config.ProviderCode))
	config.AccountID = strings.TrimSpace(config.AccountID)
	config.Jurisdiction = strings.ToLower(strings.TrimSpace(config.Jurisdiction))
	config.AccessKeyID = strings.TrimSpace(config.AccessKeyID)
	config.SecretAccessKey = strings.TrimSpace(config.SecretAccessKey)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.Region = strings.TrimSpace(config.Region)
	config.AppID = strings.TrimSpace(config.AppID)
	config.SecretID = strings.TrimSpace(config.SecretID)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	config.Endpoint = strings.TrimRight(strings.TrimSpace(config.Endpoint), "/")
	config.PublicBaseURL = strings.TrimRight(strings.TrimSpace(config.PublicBaseURL), "/")

	switch config.ProviderCode {
	case ProviderLocal:
		return config, nil
	case ProviderCloudflareR2:
		if config.AccountID == "" {
			return ConnectionConfig{}, errors.New("account_id is required")
		}
		if config.AccessKeyID == "" {
			return ConnectionConfig{}, errors.New("access_key_id is required")
		}
		if config.SecretAccessKey == "" {
			return ConnectionConfig{}, errors.New("secret_access_key is required")
		}
		if config.Bucket == "" {
			return ConnectionConfig{}, errors.New("bucket is required")
		}
		if config.Jurisdiction != "" && config.Jurisdiction != "eu" && config.Jurisdiction != "us" && config.Jurisdiction != "fedramp" {
			return ConnectionConfig{}, errors.New("invalid jurisdiction")
		}
		config.Region = "auto"
		if config.Endpoint == "" {
			config.Endpoint = "https://" + config.AccountID + ".r2.cloudflarestorage.com"
			if config.Jurisdiction != "" {
				config.Endpoint = "https://" + config.AccountID + "." + config.Jurisdiction + ".r2.cloudflarestorage.com"
			}
		}
	case ProviderAliyunOSS:
		if config.AccessKeyID == "" {
			return ConnectionConfig{}, errors.New("access_key_id is required")
		}
		if config.SecretAccessKey == "" {
			return ConnectionConfig{}, errors.New("access_key_secret is required")
		}
		if config.Bucket == "" {
			return ConnectionConfig{}, errors.New("bucket is required")
		}
		if config.Region == "" {
			return ConnectionConfig{}, errors.New("region is required")
		}
		if config.Endpoint == "" {
			config.Endpoint = "https://oss-" + config.Region + ".aliyuncs.com"
		}
		if config.PublicBaseURL == "" {
			config.PublicBaseURL = "https://" + config.Bucket + ".oss-" + config.Region + ".aliyuncs.com"
		}
	case ProviderTencentCOS:
		if config.AppID == "" {
			return ConnectionConfig{}, errors.New("app_id is required")
		}
		if config.SecretID == "" {
			return ConnectionConfig{}, errors.New("secret_id is required")
		}
		if config.SecretKey == "" {
			return ConnectionConfig{}, errors.New("secret_key is required")
		}
		if config.Bucket == "" {
			return ConnectionConfig{}, errors.New("bucket is required")
		}
		if config.Region == "" {
			return ConnectionConfig{}, errors.New("region is required")
		}
		if !strings.HasSuffix(config.Bucket, "-"+config.AppID) {
			config.Bucket += "-" + config.AppID
		}
		if config.Endpoint == "" {
			config.Endpoint = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, config.Region)
		}
		if config.PublicBaseURL == "" {
			config.PublicBaseURL = config.Endpoint
		}
	default:
		return ConnectionConfig{}, ErrUnsupportedProvider
	}

	if config.Endpoint != "" {
		if err := validateEndpoint(config.Endpoint); err != nil {
			return ConnectionConfig{}, err
		}
	}
	if config.PublicBaseURL != "" {
		if err := validateEndpoint(config.PublicBaseURL); err != nil {
			return ConnectionConfig{}, err
		}
	}
	return config, nil
}

func validateEndpoint(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ErrInvalidEndpoint
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return ErrInvalidEndpoint
	}
	return nil
}
