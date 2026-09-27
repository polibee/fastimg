package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"goravel/app/facades"
	"goravel/app/models"
	settingsservices "goravel/app/services/settings"
)

const configuredSecret = "__configured__"

var (
	ErrStorageConnectionNotFound = errors.New("storage connection not found")
	ErrStorageConnectionInvalid  = errors.New("invalid storage connection")
	ErrProviderAdapterPending    = errors.New("storage provider adapter is pending")
)

type ConnectionService struct{}

func NewConnectionService() *ConnectionService { return &ConnectionService{} }

type ConfigField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Secret      bool   `json:"secret"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
}

type ProviderDefinition struct {
	Code             string        `json:"code"`
	Label            string        `json:"label"`
	Description      string        `json:"description"`
	RegistrationURL  string        `json:"registration_url,omitempty"`
	AdapterAvailable bool          `json:"adapter_available"`
	Fields           []ConfigField `json:"fields"`
}

type ConnectionView struct {
	ID                  uint              `json:"id"`
	ProviderCode        string            `json:"provider_code"`
	Name                string            `json:"name"`
	Enabled             bool              `json:"enabled"`
	IsPrimary           bool              `json:"is_primary"`
	Status              string            `json:"status"`
	PublicBaseURL       string            `json:"public_base_url"`
	PathPrefix          string            `json:"path_prefix"`
	DefaultVisibility   string            `json:"default_visibility"`
	SignedURLTTLSeconds int               `json:"signed_url_ttl_seconds"`
	LastCheckedAt       *time.Time        `json:"last_checked_at,omitempty"`
	LastSuccessAt       *time.Time        `json:"last_success_at,omitempty"`
	LastErrorCode       string            `json:"last_error_code,omitempty"`
	LastErrorMessage    string            `json:"last_error_message,omitempty"`
	Config              map[string]string `json:"config"`
}

type ConnectionInput struct {
	Name                string            `json:"name"`
	Enabled             bool              `json:"enabled"`
	IsPrimary           bool              `json:"is_primary"`
	PublicBaseURL       string            `json:"public_base_url"`
	PathPrefix          string            `json:"path_prefix"`
	DefaultVisibility   string            `json:"default_visibility"`
	SignedURLTTLSeconds int               `json:"signed_url_ttl_seconds"`
	Config              map[string]string `json:"config"`
}

func ProviderDefinitions() map[string]*ProviderDefinition {
	return map[string]*ProviderDefinition{
		ProviderLocal: {
			Code: ProviderLocal, Label: "Local", Description: "Store media on the application server.", AdapterAvailable: true,
		},
		ProviderCloudflareR2: {
			Code: ProviderCloudflareR2, Label: "Cloudflare R2", Description: "S3-compatible object storage with an Account ID and R2 API token.", RegistrationURL: "https://dash.cloudflare.com/", AdapterAvailable: true,
			Fields: []ConfigField{
				{Name: "account_id", Label: "Account ID", Required: true},
				{Name: "jurisdiction", Label: "Jurisdiction"},
				{Name: "access_key_id", Label: "Access Key ID", Secret: true, Required: true},
				{Name: "secret_access_key", Label: "Secret Access Key", Secret: true, Required: true},
				{Name: "bucket", Label: "Bucket", Required: true},
				{Name: "endpoint", Label: "S3 Endpoint", Placeholder: "https://<account>.r2.cloudflarestorage.com"},
			},
		},
		ProviderAliyunOSS: {
			Code: ProviderAliyunOSS, Label: "Alibaba Cloud OSS", Description: "Alibaba Cloud Object Storage Service.", RegistrationURL: "https://www.aliyun.com/minisite/goods?userCode=i7hvp048", AdapterAvailable: true,
			Fields: []ConfigField{
				{Name: "access_key_id", Label: "AccessKey ID", Secret: true, Required: true},
				{Name: "access_key_secret", Label: "AccessKey Secret", Secret: true, Required: true},
				{Name: "bucket", Label: "Bucket", Required: true},
				{Name: "region", Label: "Region", Required: true, Placeholder: "cn-hangzhou"},
				{Name: "endpoint", Label: "Endpoint", Placeholder: "https://oss-cn-hangzhou.aliyuncs.com"},
			},
		},
		ProviderTencentCOS: {
			Code: ProviderTencentCOS, Label: "Tencent COS", Description: "Tencent Cloud Object Storage.", RegistrationURL: "https://curl.qcloud.com/sTyZPtN7", AdapterAvailable: true,
			Fields: []ConfigField{
				{Name: "app_id", Label: "APPID", Required: true},
				{Name: "secret_id", Label: "SecretId", Secret: true, Required: true},
				{Name: "secret_key", Label: "SecretKey", Secret: true, Required: true},
				{Name: "bucket", Label: "Bucket", Required: true},
				{Name: "region", Label: "Region", Required: true, Placeholder: "ap-guangzhou"},
				{Name: "endpoint", Label: "Endpoint", Placeholder: "https://<bucket>-<appid>.cos.<region>.myqcloud.com"},
			},
		},
	}
}

func (s *ConnectionService) List() ([]ConnectionView, error) {
	var rows []models.StorageConnection
	if err := facades.Orm().Query().OrderBy("provider_code").OrderBy("id").Get(&rows); err != nil {
		return nil, err
	}
	byProvider := make(map[string]models.StorageConnection, len(rows))
	for _, row := range rows {
		byProvider[row.ProviderCode] = row
	}
	definitions := ProviderDefinitions()
	views := make([]ConnectionView, 0, len(definitions))
	for _, code := range []string{ProviderLocal, ProviderCloudflareR2, ProviderAliyunOSS, ProviderTencentCOS} {
		definition := definitions[code]
		row, ok := byProvider[code]
		if !ok {
			row = models.StorageConnection{ProviderCode: code, Name: definition.Label, Status: models.StorageConnectionDisabled, DefaultVisibility: "public", SignedURLTTLSeconds: 3600}
		}
		view, err := connectionView(row)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *ConnectionService) Save(providerCode string, input ConnectionInput) (ConnectionView, error) {
	providerCode = strings.ToLower(strings.TrimSpace(providerCode))
	definition, ok := ProviderDefinitions()[providerCode]
	if !ok {
		return ConnectionView{}, ErrStorageConnectionInvalid
	}
	if input.IsPrimary {
		if err := validatePrimarySelection(providerCode, input.Enabled); err != nil {
			return ConnectionView{}, err
		}
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = definition.Label
	}
	var existing models.StorageConnection
	err := facades.Orm().Query().Where("provider_code = ?", providerCode).OrderBy("id").First(&existing)
	if err != nil {
		existing = models.StorageConnection{ProviderCode: providerCode, Name: name, DefaultVisibility: "public", SignedURLTTLSeconds: 3600}
	}
	current, err := decodeConfig(existing.ConfigEncrypted)
	if err != nil {
		return ConnectionView{}, err
	}
	current.ProviderCode = providerCode
	merged, err := configFromValues(current, input.Config)
	if err != nil {
		return ConnectionView{}, err
	}
	normalized, err := NormalizeConnectionConfig(merged)
	if err != nil && input.Enabled {
		return ConnectionView{}, fmt.Errorf("%w: %v", ErrStorageConnectionInvalid, err)
	}
	if err != nil {
		normalized = merged
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return ConnectionView{}, err
	}
	ciphertext, err := settingsservices.EncryptSecret(string(encoded))
	if err != nil {
		return ConnectionView{}, err
	}
	status, statusCode := statusForProvider(providerCode, input.Enabled, err)
	if !definition.AdapterAvailable && input.Enabled && err == nil {
		status, statusCode = models.StorageConnectionDegraded, "STORAGE_PROVIDER_ADAPTER_PENDING"
	}
	publicURL := strings.TrimRight(strings.TrimSpace(input.PublicBaseURL), "/")
	if publicURL == "" {
		publicURL = normalized.PublicBaseURL
	}
	values := map[string]any{
		"provider_code": providerCode, "name": name, "enabled": input.Enabled, "is_primary": input.IsPrimary,
		"status": status, "config_encrypted": ciphertext, "public_base_url": nullableString(publicURL),
		"path_prefix": nullableString(strings.Trim(strings.TrimSpace(input.PathPrefix), "/")), "default_visibility": defaultVisibility(input.DefaultVisibility),
		"signed_url_ttl_seconds": signedURLTTL(input.SignedURLTTLSeconds), "last_error_code": nullableString(statusCode),
		"last_error_message": nullableString(adapterStatusMessage(statusCode)), "updated_at": time.Now(),
	}
	if input.Enabled {
		now := time.Now()
		values["last_checked_at"] = now
		if status == models.StorageConnectionHealthy {
			values["last_success_at"] = now
		}
	}
	if existing.ID == 0 {
		if err := facades.Orm().Query().Create(&values); err != nil {
			return ConnectionView{}, err
		}
		if err := facades.Orm().Query().Where("provider_code = ? AND name = ?", providerCode, name).First(&existing); err != nil {
			return ConnectionView{}, err
		}
	} else {
		if _, err := facades.Orm().Query().Table("storage_connections").Where("id = ?", existing.ID).Update(values); err != nil {
			return ConnectionView{}, err
		}
	}
	if input.IsPrimary && input.Enabled {
		if _, err := facades.Orm().Query().Table("storage_connections").Where("id <> ?", existing.ID).Update(map[string]any{"is_primary": false, "updated_at": time.Now()}); err != nil {
			return ConnectionView{}, err
		}
	}
	if err := facades.Orm().Query().Where("id = ?", existing.ID).First(&existing); err != nil {
		return ConnectionView{}, err
	}
	return connectionView(existing)
}

func (s *ConnectionService) Test(providerCode string) (ConnectionView, error) {
	providerCode = strings.ToLower(strings.TrimSpace(providerCode))
	var row models.StorageConnection
	if err := facades.Orm().Query().Where("provider_code = ?", providerCode).OrderBy("id").First(&row); err != nil {
		return ConnectionView{}, ErrStorageConnectionNotFound
	}
	config, err := decodeConfig(row.ConfigEncrypted)
	if err != nil {
		return ConnectionView{}, err
	}
	if _, err := NormalizeConnectionConfig(config); err != nil {
		return ConnectionView{}, fmt.Errorf("%w: %v", ErrStorageConnectionInvalid, err)
	}
	status, code := statusForProvider(providerCode, true, nil)
	if providerCode != ProviderLocal {
		providerConfig, providerErr := newProviderFromConfig(normalizedConfig(config), nil)
		if providerErr == nil {
			_, providerErr = providerConfig.Exists(context.Background(), ".fastimg/healthcheck")
		}
		if providerErr != nil {
			status = models.StorageConnectionError
			code = "STORAGE_CONNECTION_TEST_FAILED"
		} else {
			status = models.StorageConnectionHealthy
			code = ""
		}
	}
	now := time.Now()
	values := map[string]any{"status": status, "last_checked_at": now, "last_error_code": nullableString(code), "last_error_message": nullableString(adapterStatusMessage(code)), "updated_at": now}
	if status == models.StorageConnectionHealthy {
		values["last_success_at"] = now
	}
	if _, err := facades.Orm().Query().Table("storage_connections").Where("id = ?", row.ID).Update(values); err != nil {
		return ConnectionView{}, err
	}
	if err := facades.Orm().Query().Where("id = ?", row.ID).First(&row); err != nil {
		return ConnectionView{}, err
	}
	return connectionView(row)
}

func normalizedConfig(config ConnectionConfig) ConnectionConfig {
	normalized, err := NormalizeConnectionConfig(config)
	if err != nil {
		return config
	}
	return normalized
}

func connectionView(row models.StorageConnection) (ConnectionView, error) {
	config, err := decodeConfig(row.ConfigEncrypted)
	if err != nil {
		return ConnectionView{}, err
	}
	return ConnectionView{
		ID: row.ID, ProviderCode: row.ProviderCode, Name: row.Name, Enabled: row.Enabled, IsPrimary: row.IsPrimary, Status: row.Status,
		PublicBaseURL: row.PublicBaseURL, PathPrefix: row.PathPrefix, DefaultVisibility: row.DefaultVisibility, SignedURLTTLSeconds: row.SignedURLTTLSeconds,
		LastCheckedAt: row.LastCheckedAt, LastSuccessAt: row.LastSuccessAt, LastErrorCode: row.LastErrorCode, LastErrorMessage: row.LastErrorMessage,
		Config: redactConfig(config),
	}, nil
}

func decodeConfig(value string) (ConnectionConfig, error) {
	if strings.TrimSpace(value) == "" {
		return ConnectionConfig{}, nil
	}
	plaintext, err := settingsservices.DecryptSecret(value)
	if err != nil {
		return ConnectionConfig{}, err
	}
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(plaintext), &config); err != nil {
		return ConnectionConfig{}, err
	}
	return config, nil
}

func redactConfig(config ConnectionConfig) map[string]string {
	values := map[string]string{"provider_code": config.ProviderCode, "account_id": config.AccountID, "jurisdiction": config.Jurisdiction, "bucket": config.Bucket, "region": config.Region, "app_id": config.AppID, "endpoint": config.Endpoint, "public_base_url": config.PublicBaseURL}
	if config.AccessKeyID != "" {
		values["access_key_id"] = configuredSecret
	}
	if config.SecretAccessKey != "" {
		if config.ProviderCode == ProviderAliyunOSS {
			values["access_key_secret"] = configuredSecret
		} else {
			values["secret_access_key"] = configuredSecret
		}
	}
	if config.SecretID != "" {
		values["secret_id"] = configuredSecret
	}
	if config.SecretKey != "" {
		values["secret_key"] = configuredSecret
	}
	return values
}

func configFromValues(existing ConnectionConfig, values map[string]string) (ConnectionConfig, error) {
	config := existing
	for key, value := range values {
		value = strings.TrimSpace(value)
		switch key {
		case "provider_code":
			config.ProviderCode = value
		case "account_id":
			config.AccountID = value
		case "jurisdiction":
			config.Jurisdiction = value
		case "access_key_id":
			if value != "" && value != configuredSecret {
				config.AccessKeyID = value
			}
		case "secret_access_key":
			if value != "" && value != configuredSecret {
				config.SecretAccessKey = value
			}
		case "access_key_secret":
			if value != "" && value != configuredSecret {
				config.SecretAccessKey = value
			}
		case "bucket":
			config.Bucket = value
		case "region":
			config.Region = value
		case "app_id":
			config.AppID = value
		case "secret_id":
			if value != "" && value != configuredSecret {
				config.SecretID = value
			}
		case "secret_key":
			if value != "" && value != configuredSecret {
				config.SecretKey = value
			}
		case "endpoint":
			config.Endpoint = value
		case "public_base_url":
			config.PublicBaseURL = value
		}
	}
	return config, nil
}

func statusForProvider(providerCode string, enabled bool, validationErr error) (string, string) {
	if !enabled {
		return models.StorageConnectionDisabled, ""
	}
	if validationErr != nil {
		return models.StorageConnectionIncomplete, "STORAGE_CONFIGURATION_INVALID"
	}
	if providerCode == ProviderLocal {
		return models.StorageConnectionHealthy, ""
	}
	return models.StorageConnectionDegraded, "STORAGE_CONNECTION_UNTESTED"
}

func validatePrimarySelection(providerCode string, enabled bool) error {
	if !enabled {
		return ErrStorageConnectionInvalid
	}
	definition, ok := ProviderDefinitions()[providerCode]
	if !ok {
		return ErrStorageConnectionInvalid
	}
	if !definition.AdapterAvailable {
		return ErrProviderAdapterPending
	}
	return nil
}

func adapterStatusMessage(code string) string {
	if code == "STORAGE_PROVIDER_ADAPTER_PENDING" {
		return "Provider SDK adapter is not enabled in this build yet."
	}
	if code == "STORAGE_CONFIGURATION_INVALID" {
		return "Provider credentials or endpoint are incomplete."
	}
	if code == "STORAGE_CONNECTION_TEST_FAILED" {
		return "The provider could not be reached with the saved credentials."
	}
	if code == "STORAGE_CONNECTION_UNTESTED" {
		return "The provider is configured. Run a connection test before relying on it for uploads."
	}
	return ""
}

func defaultVisibility(value string) string {
	if value != "private" {
		return "public"
	}
	return value
}

func signedURLTTL(value int) int {
	if value < 60 || value > 86400 {
		return 3600
	}
	return value
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
