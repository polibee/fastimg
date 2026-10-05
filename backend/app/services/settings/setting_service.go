package settingsservices

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"goravel/app/facades"
	"goravel/app/models"
)

var (
	ErrInvalidSetting = errors.New("invalid system setting")
)

var settingKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,159}$`)

type SettingService struct{}

func NewSettingService() *SettingService { return &SettingService{} }

func normalizeSettingType(valueType string) string {
	valueType = strings.ToLower(strings.TrimSpace(valueType))
	if valueType == "" || valueType == "select" {
		return "string"
	}
	return valueType
}

func validateSetting(key, value, valueType string) error {
	key = strings.TrimSpace(key)
	valueType = normalizeSettingType(valueType)
	if !settingKeyPattern.MatchString(key) {
		return ErrInvalidSetting
	}
	switch valueType {
	case "string", "secret":
		return nil
	case "boolean":
		if value == "true" || value == "false" {
			return nil
		}
	case "integer":
		// Optional numeric settings (for example statistics retention) may be
		// left blank in the admin form. Consumers apply their own safe default.
		if strings.TrimSpace(value) == "" {
			return nil
		}
		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			return nil
		}
	case "json":
		if json.Valid([]byte(value)) {
			return nil
		}
	}
	return ErrInvalidSetting
}

func (s *SettingService) List() ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	// `group` is a PostgreSQL reserved keyword; ordering by it unquoted makes
	// the otherwise valid settings endpoint fail with a generic 500.
	if err := facades.Orm().Query().OrderBy("key").Get(&settings); err != nil {
		return nil, err
	}
	for index := range settings {
		if IsSecretKey(settings[index].Key) && strings.TrimSpace(settings[index].Value) != "" {
			settings[index].Value = secretPlaceholder
		}
	}
	return settings, nil
}

// Resolve returns a setting value for trusted backend consumers. Provider
// credentials are decrypted here and are never returned by List or Upsert.
func (s *SettingService) Resolve(key, fallback string) string {
	var setting models.SystemSetting
	if !facades.Schema().HasTable("system_settings") {
		return fallback
	}
	if err := facades.Orm().Query().Where("key = ?", key).First(&setting); err != nil || strings.TrimSpace(setting.Value) == "" {
		return fallback
	}
	if IsSecretKey(key) {
		value, err := decryptSecret(setting.Value)
		if err != nil {
			return fallback
		}
		return value
	}
	return strings.TrimSpace(setting.Value)
}

func (s *SettingService) Upsert(key, value, valueType, group, description string) (*models.SystemSetting, error) {
	key = strings.TrimSpace(key)
	valueType = normalizeSettingType(valueType)
	group = strings.TrimSpace(group)
	if group == "" {
		group = "general"
	}
	if err := validateSetting(key, value, valueType); err != nil {
		return nil, err
	}

	var existing []models.SystemSetting
	if err := facades.Orm().Query().Where("key = ?", key).Get(&existing); err != nil {
		return nil, err
	}
	storedValue := value
	if IsSecretKey(key) {
		if value == secretPlaceholder && len(existing) > 0 && strings.TrimSpace(existing[0].Value) != "" {
			storedValue = existing[0].Value
		} else if value != "" {
			var err error
			storedValue, err = encryptSecret(value)
			if err != nil {
				return nil, ErrInvalidSetting
			}
		}
	}
	values := map[string]any{
		"key":         key,
		"value":       storedValue,
		"value_type":  valueType,
		"group":       group,
		"description": strings.TrimSpace(description),
	}
	if len(existing) == 0 {
		if err := facades.Orm().Query().Table("system_settings").Create(&values); err != nil {
			return nil, err
		}
	} else if _, err := facades.Orm().Query().Table("system_settings").Where("key = ?", key).Update(values); err != nil {
		return nil, err
	}

	var setting models.SystemSetting
	if err := facades.Orm().Query().Where("key = ?", key).First(&setting); err != nil {
		return nil, err
	}
	if IsSecretKey(setting.Key) && strings.TrimSpace(setting.Value) != "" {
		setting.Value = secretPlaceholder
	}
	return &setting, nil
}
