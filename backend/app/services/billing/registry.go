package billing

import (
	"sync"

	"goravel/app/services/billing/providers"
	"goravel/app/services/billing/providers/fake"
)

var (
	developmentRegistryOnce sync.Once
	developmentRegistry     *providers.Registry
)

// DefaultGatewayRegistry is process-scoped so member checkout and webhook
// handling use the same Fake Provider instance during local development.
func DefaultGatewayRegistry() *providers.Registry {
	developmentRegistryOnce.Do(func() {
		developmentRegistry = providers.NewRegistry()
		_ = developmentRegistry.Register("fake", fake.New())
	})
	return developmentRegistry
}
