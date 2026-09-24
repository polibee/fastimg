package providers

import "errors"

var ErrGatewayUnavailable = errors.New("payment gateway unavailable")

type Registry struct{ gateways map[string]PaymentGateway }

func NewRegistry() *Registry { return &Registry{gateways: map[string]PaymentGateway{}} }

func (r *Registry) Register(code string, gateway PaymentGateway) error {
	if code == "" || gateway == nil {
		return ErrGatewayUnavailable
	}
	r.gateways[code] = gateway
	return nil
}

func (r *Registry) Get(code string) (PaymentGateway, error) {
	gateway, ok := r.gateways[code]
	if !ok {
		return nil, ErrGatewayUnavailable
	}
	return gateway, nil
}
