package storage

import (
	"errors"
	"strings"

	"goravel/app/facades"
	"goravel/app/models"
)

var (
	ErrStorageConnectionDisabled = errors.New("storage connection is disabled")
	ErrStorageConnectionMissing  = errors.New("storage connection is missing")
)

type RuntimeRegistry struct {
	disk localDisk
}

func NewRuntimeRegistry(disk localDisk) *RuntimeRegistry {
	return &RuntimeRegistry{disk: disk}
}

// ProviderFor resolves a persisted connection into the business-level storage
// interface. Provider-specific SDKs are deliberately hidden behind this
// boundary; callers never switch on cloud vendor names.
func (r *RuntimeRegistry) ProviderFor(connection models.StorageConnection) (StorageProvider, error) {
	if !connection.Enabled {
		return nil, ErrStorageConnectionDisabled
	}
	config, err := decodeConfig(connection.ConfigEncrypted)
	if err != nil {
		return nil, err
	}
	config.ProviderCode = strings.ToLower(strings.TrimSpace(connection.ProviderCode))
	if _, err := NormalizeConnectionConfig(config); err != nil {
		return nil, err
	}
	switch config.ProviderCode {
	case ProviderLocal:
		return NewLocalProvider(r.disk), nil
	case ProviderCloudflareR2, ProviderAliyunOSS, ProviderTencentCOS:
		return newProviderFromConfig(config, nil)
	default:
		return nil, ErrUnsupportedProvider
	}
}

func (r *RuntimeRegistry) Primary() (StorageProvider, models.StorageConnection, error) {
	var connection models.StorageConnection
	if err := facades.Orm().Query().Where("enabled = ? AND is_primary = ?", true, true).OrderBy("id").First(&connection); err != nil {
		if err := facades.Orm().Query().Where("provider_code = ? AND enabled = ?", ProviderLocal, true).OrderBy("id").First(&connection); err != nil {
			return nil, models.StorageConnection{}, ErrStorageConnectionMissing
		}
	}
	provider, err := r.ProviderFor(connection)
	if err != nil {
		return nil, connection, err
	}
	return provider, connection, nil
}
