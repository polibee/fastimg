package storage

import (
	"testing"

	"goravel/app/models"
)

func TestRuntimeRegistryResolvesEnabledLocalConnection(t *testing.T) {
	provider, err := NewRuntimeRegistry(newMemoryDisk()).ProviderFor(models.StorageConnection{
		ProviderCode:    ProviderLocal,
		Enabled:         true,
		ConfigEncrypted: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*LocalProvider); !ok {
		t.Fatalf("provider type = %T, want *LocalProvider", provider)
	}
}
