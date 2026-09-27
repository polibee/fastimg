package providers

import (
	"errors"
	"sort"
	"sync"
)

var ErrGatewayUnavailable = errors.New("payment gateway unavailable")

type Registry struct {
	mu       sync.RWMutex
	gateways map[string]PaymentGateway
}

func NewRegistry() *Registry { return &Registry{gateways: map[string]PaymentGateway{}} }

func (r *Registry) Register(code string, gateway PaymentGateway) error {
	if code == "" || gateway == nil {
		return ErrGatewayUnavailable
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gateways[code] = gateway
	return nil
}

// Replace swaps the configured providers without replacing the registry
// pointer. Controllers and services keep the same registry while settings can
// be reloaded safely after an administrator saves gateway configuration.
func (r *Registry) Replace(gateways map[string]PaymentGateway) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gateways = make(map[string]PaymentGateway, len(gateways))
	for code, gateway := range gateways {
		r.gateways[code] = gateway
	}
}

func (r *Registry) Get(code string) (PaymentGateway, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gateway, ok := r.gateways[code]
	if !ok {
		return nil, ErrGatewayUnavailable
	}
	return gateway, nil
}

func (r *Registry) Codes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	codes := make([]string, 0, len(r.gateways))
	for code := range r.gateways {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}
