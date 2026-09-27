package paypal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"goravel/app/services/billing/providers"
)

var ErrInvalidResponse = errors.New("invalid paypal response")
var ErrUnsupportedRefund = errors.New("paypal refund is not available for this payment")
var ErrInvalidWebhook = errors.New("invalid paypal webhook")

type Config struct {
	Enabled      bool
	Environment  string
	BaseURL      string
	ClientID     string
	ClientSecret string
	WebhookID    string
	Timeout      time.Duration
	ReturnURL    string
	CancelURL    string
}

type Provider struct {
	config         Config
	client         *http.Client
	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

func New(config Config) *Provider {
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	return &Provider{config: config, client: &http.Client{Timeout: config.Timeout}}
}

func (p *Provider) CreatePayment(ctx context.Context, request providers.CreatePaymentRequest) (providers.PaymentSession, error) {
	if !p.ready() || request.OrderNo == "" || request.AmountMinor <= 0 {
		return providers.PaymentSession{}, providers.ErrGatewayUnavailable
	}
	payload := map[string]any{"intent": "CAPTURE", "purchase_units": []any{map[string]any{"reference_id": request.OrderNo, "invoice_id": request.OrderNo, "description": truncate(request.Description), "amount": map[string]string{"currency_code": request.Currency, "value": formatMinor(request.AmountMinor)}}}, "application_context": map[string]string{"return_url": orderReturnURL(p.config.ReturnURL, request.OrderNo), "cancel_url": orderReturnURL(p.config.CancelURL, request.OrderNo)}}
	body, err := json.Marshal(payload)
	if err != nil {
		return providers.PaymentSession{}, err
	}
	responseBody, _, err := p.request(ctx, http.MethodPost, "/v2/checkout/orders", body, request.IdempotencyKey)
	if err != nil {
		return providers.PaymentSession{}, err
	}
	var response orderResponse
	if json.Unmarshal(responseBody, &response) != nil || response.ID == "" {
		return providers.PaymentSession{}, ErrInvalidResponse
	}
	return providers.PaymentSession{GatewayCode: "paypal", ProviderPaymentID: response.ID, ProviderOrderID: response.ID, Status: mapOrderStatus(response.Status), CheckoutURL: response.approvalURL(), OccurredAt: time.Now().Unix()}, nil
}

func orderReturnURL(template, orderNo string) string {
	return strings.ReplaceAll(template, "{order_id}", orderNo)
}

func (p *Provider) QueryPayment(ctx context.Context, request providers.QueryPaymentRequest) (providers.GatewayPayment, error) {
	if !p.ready() || request.ProviderPaymentID == "" {
		return providers.GatewayPayment{}, providers.ErrGatewayUnavailable
	}
	body, _, err := p.request(ctx, http.MethodGet, "/v2/checkout/orders/"+segment(request.ProviderPaymentID), nil, "")
	if err != nil {
		return providers.GatewayPayment{}, err
	}
	var response orderResponse
	if json.Unmarshal(body, &response) != nil || response.ID == "" {
		return providers.GatewayPayment{}, ErrInvalidResponse
	}
	amount, currency, captureID := response.captureDetails()
	if amount == 0 || currency == "" {
		return providers.GatewayPayment{}, ErrInvalidResponse
	}
	return providers.GatewayPayment{GatewayCode: "paypal", ProviderPaymentID: response.ID, ProviderPaymentTx: captureID, Status: mapOrderStatus(response.Status), AmountMinor: amount, Currency: currency, OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) VerifyWebhook(ctx context.Context, request providers.WebhookRequest) (providers.GatewayEvent, error) {
	if !p.ready() || p.config.WebhookID == "" {
		return providers.GatewayEvent{}, providers.ErrGatewayUnavailable
	}
	verification := map[string]any{"transmission_id": header(request.Headers, "paypal-transmission-id"), "transmission_time": header(request.Headers, "paypal-transmission-time"), "cert_url": header(request.Headers, "paypal-cert-url"), "auth_algo": header(request.Headers, "paypal-auth-algo"), "transmission_sig": header(request.Headers, "paypal-transmission-sig"), "webhook_id": p.config.WebhookID, "webhook_event": json.RawMessage(request.Body)}
	body, _ := json.Marshal(verification)
	verified, _, err := p.request(ctx, http.MethodPost, "/v1/notifications/verify-webhook-signature", body, "")
	if err != nil {
		return providers.GatewayEvent{}, err
	}
	var result struct {
		Status string `json:"verification_status"`
	}
	if json.Unmarshal(verified, &result) != nil || result.Status != "SUCCESS" {
		return providers.GatewayEvent{}, ErrInvalidWebhook
	}
	var event webhookEvent
	if json.Unmarshal(request.Body, &event) != nil || event.ID == "" {
		return providers.GatewayEvent{}, ErrInvalidResponse
	}
	amount := parseMinor(event.Resource.Amount.Value)
	currency := event.Resource.Amount.Currency
	if amount == 0 || currency == "" {
		return providers.GatewayEvent{}, ErrInvalidResponse
	}
	providerID := event.Resource.Supplementary.Related.OrderID
	if providerID == "" {
		providerID = event.Resource.ID
	}
	return providers.GatewayEvent{GatewayCode: "paypal", EventID: event.ID, EventType: event.Type, ProviderPaymentID: providerID, ProviderPaymentTx: event.Resource.ID, Status: mapEventStatus(event.Type, event.Resource.Status), AmountMinor: amount, Currency: currency, OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) CreateRefund(ctx context.Context, request providers.RefundRequest) (providers.GatewayRefund, error) {
	if !p.ready() || request.ProviderPaymentID == "" {
		return providers.GatewayRefund{}, providers.ErrGatewayUnavailable
	}
	payment, err := p.QueryPayment(ctx, providers.QueryPaymentRequest{ProviderPaymentID: request.ProviderPaymentID})
	if err != nil || payment.ProviderPaymentTx == "" {
		return providers.GatewayRefund{}, ErrUnsupportedRefund
	}
	payload := []byte(`{}`)
	if request.AmountMinor > 0 {
		payload = []byte(fmt.Sprintf(`{"amount":{"value":"%s","currency_code":"%s"}}`, formatMinor(request.AmountMinor), request.Currency))
	}
	body, _, err := p.request(ctx, http.MethodPost, "/v2/payments/captures/"+segment(payment.ProviderPaymentTx)+"/refund", payload, request.IdempotencyKey)
	if err != nil {
		return providers.GatewayRefund{}, err
	}
	var response refundResponse
	if json.Unmarshal(body, &response) != nil || response.ID == "" {
		return providers.GatewayRefund{}, ErrInvalidResponse
	}
	amount := request.AmountMinor
	if amount == 0 {
		amount = payment.AmountMinor
	}
	return providers.GatewayRefund{GatewayCode: "paypal", ProviderRefundID: response.ID, Status: mapRefundStatus(response.Status), AmountMinor: amount, Currency: request.Currency, OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) QueryRefund(context.Context, string) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, ErrUnsupportedRefund
}

func (p *Provider) ready() bool {
	return p != nil && p.config.Enabled && p.config.BaseURL != "" && p.config.ClientID != "" && p.config.ClientSecret != ""
}

func (p *Provider) accessToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.token != "" && time.Now().Before(p.tokenExpiresAt) {
		return p.token, nil
	}
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(p.config.ClientID+":"+p.config.ClientSecret)))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := p.client.Do(request)
	if err != nil {
		return "", &providers.RequestError{GatewayCode: "paypal", Retryable: true, ProviderMessage: "network request failed"}
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", parseRequestError(response.StatusCode, body)
	}
	var token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if json.Unmarshal(body, &token) != nil || token.AccessToken == "" {
		return "", ErrInvalidResponse
	}
	p.token = token.AccessToken
	ttl := time.Duration(token.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	p.tokenExpiresAt = time.Now().Add(ttl - 30*time.Second)
	return p.token, nil
}

func (p *Provider) request(ctx context.Context, method, path string, body []byte, requestID string) ([]byte, int, error) {
	token, err := p.accessToken(ctx)
	if err != nil {
		return nil, 0, err
	}
	request, err := http.NewRequestWithContext(ctx, method, p.config.BaseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		request.Header.Set("PayPal-Request-Id", requestID)
	}
	response, err := p.client.Do(request)
	if err != nil {
		return nil, 0, &providers.RequestError{GatewayCode: "paypal", Retryable: true, ProviderMessage: "network request failed"}
	}
	defer response.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20+1))
	if readErr != nil || len(data) > 1<<20 {
		return nil, response.StatusCode, ErrInvalidResponse
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, response.StatusCode, parseRequestError(response.StatusCode, data)
	}
	return data, response.StatusCode, nil
}

func parseRequestError(status int, body []byte) error {
	var payload struct {
		Name    string `json:"name"`
		Message string `json:"message"`
		Error   string `json:"error_description"`
	}
	providerCode, providerMessage := "", ""
	if json.Unmarshal(body, &payload) == nil {
		providerCode = strings.TrimSpace(payload.Name)
		providerMessage = strings.TrimSpace(payload.Message)
		if providerMessage == "" {
			providerMessage = strings.TrimSpace(payload.Error)
		}
	}
	return &providers.RequestError{
		GatewayCode: "paypal", StatusCode: status, ProviderCode: providerCode,
		ProviderMessage: providerMessage, Retryable: status == http.StatusTooManyRequests || status >= 500,
	}
}

type orderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					Currency string `json:"currency_code"`
					Value    string `json:"value"`
				} `json:"amount"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

func (r orderResponse) approvalURL() string {
	for _, link := range r.Links {
		if link.Rel == "approve" {
			return link.Href
		}
	}
	return ""
}
func (r orderResponse) captureDetails() (int64, string, string) {
	for _, unit := range r.PurchaseUnits {
		for _, capture := range unit.Payments.Captures {
			amount := parseMinor(capture.Amount.Value)
			return amount, capture.Amount.Currency, capture.ID
		}
	}
	return 0, "", ""
}

type webhookEvent struct {
	ID       string `json:"id"`
	Type     string `json:"event_type"`
	Resource struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Currency string `json:"currency_code"`
			Value    string `json:"value"`
		} `json:"amount"`
		Supplementary struct {
			Related struct {
				OrderID string `json:"order_id"`
			} `json:"related_ids"`
		} `json:"supplementary_data"`
	} `json:"resource"`
}
type refundResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func mapOrderStatus(status string) string {
	switch status {
	case "COMPLETED":
		return "succeeded"
	case "APPROVED", "CREATED":
		return "requires_action"
	case "VOIDED", "CANCELED":
		return "failed"
	default:
		return "pending"
	}
}
func mapEventStatus(eventType, status string) string {
	if eventType == "PAYMENT.CAPTURE.COMPLETED" || status == "COMPLETED" {
		return "succeeded"
	}
	if eventType == "PAYMENT.CAPTURE.DENIED" || eventType == "CHECKOUT.PAYMENT-APPROVAL.REVERSED" {
		return "failed"
	}
	return "pending"
}
func mapRefundStatus(status string) string {
	if status == "COMPLETED" {
		return "succeeded"
	}
	if status == "FAILED" {
		return "failed"
	}
	return "pending"
}
func parseMinor(value string) int64 {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) > 2 || len(parts) == 0 {
		return 0
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 && strings.TrimRight(fraction[2:], "0") != "" {
		return 0
	}
	fraction = (fraction + "00")[:2]
	minor, _ := strconv.ParseInt(fraction, 10, 64)
	return whole*100 + minor
}
func formatMinor(value int64) string { return fmt.Sprintf("%d.%02d", value/100, value%100) }
func truncate(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 127 {
		return value[:127]
	}
	if value == "" {
		return "FastImg plan"
	}
	return value
}
func segment(value string) string { return strings.Trim(strings.ReplaceAll(value, "/", ""), " ") }
func header(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
