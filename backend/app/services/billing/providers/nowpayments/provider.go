package nowpayments

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"goravel/app/services/billing/providers"
)

var ErrInvalidResponse = errors.New("invalid nowpayments response")
var ErrUnsupportedRefund = errors.New("nowpayments refunds require manual processing")

type Config struct {
	Enabled     bool
	BaseURL     string
	APIKey      string
	IPNSecret   string
	Timeout     time.Duration
	CallbackURL string
	SuccessURL  string
	CancelURL   string
}

type Provider struct {
	config Config
	client *http.Client
}

func New(config Config) *Provider {
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	return &Provider{config: config, client: &http.Client{Timeout: config.Timeout}}
}

func (p *Provider) CreatePayment(ctx context.Context, request providers.CreatePaymentRequest) (providers.PaymentSession, error) {
	if !p.readyForAPI() || request.OrderNo == "" || request.AmountMinor <= 0 {
		return providers.PaymentSession{}, providers.ErrGatewayUnavailable
	}
	payload := map[string]any{
		"price_amount": json.Number(formatMinor(request.AmountMinor)), "price_currency": strings.ToLower(request.Currency),
		"order_id": request.OrderNo, "order_description": truncate(request.Description),
		"ipn_callback_url": p.config.CallbackURL, "success_url": orderReturnURL(p.config.SuccessURL, request.OrderNo), "cancel_url": orderReturnURL(p.config.CancelURL, request.OrderNo),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return providers.PaymentSession{}, err
	}
	responseBody, _, err := p.request(ctx, http.MethodPost, "/v1/invoice", body)
	if err != nil {
		return providers.PaymentSession{}, err
	}
	var response paymentResponse
	if err := json.Unmarshal(responseBody, &response); err != nil || response.ID == "" {
		return providers.PaymentSession{}, ErrInvalidResponse
	}
	return providers.PaymentSession{GatewayCode: "nowpayments", ProviderPaymentID: response.ID, Status: MapStatus(response.Status), CheckoutURL: response.InvoiceURL, OccurredAt: time.Now().Unix()}, nil
}

func orderReturnURL(template, orderNo string) string {
	return strings.ReplaceAll(template, "{order_id}", orderNo)
}

func (p *Provider) QueryPayment(ctx context.Context, request providers.QueryPaymentRequest) (providers.GatewayPayment, error) {
	if !p.readyForAPI() || request.ProviderPaymentID == "" {
		return providers.GatewayPayment{}, providers.ErrGatewayUnavailable
	}
	body, _, err := p.request(ctx, http.MethodGet, "/v1/payment/"+urlSegment(request.ProviderPaymentID), nil)
	if err != nil {
		return providers.GatewayPayment{}, err
	}
	var response paymentResponse
	if err := json.Unmarshal(body, &response); err != nil || response.ID == "" {
		return providers.GatewayPayment{}, ErrInvalidResponse
	}
	amount, err := parseMinor(response.PriceAmount)
	if err != nil {
		return providers.GatewayPayment{}, err
	}
	return providers.GatewayPayment{GatewayCode: "nowpayments", ProviderPaymentID: response.ID, ProviderPaymentTx: response.TxID, Status: MapStatus(response.Status), AmountMinor: amount, Currency: strings.ToUpper(response.PriceCurrency), OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) VerifyWebhook(_ context.Context, request providers.WebhookRequest) (providers.GatewayEvent, error) {
	if !p.readyForWebhook() || !verifyIPN(p.config.IPNSecret, signatureHeader(request.Headers, "x-nowpayments-sig"), request.Body) {
		return providers.GatewayEvent{}, providers.ErrGatewayUnavailable
	}
	var payload ipnPayload
	if err := json.Unmarshal(request.Body, &payload); err != nil || rawString(payload.PaymentID) == "" || payload.OrderID == "" {
		return providers.GatewayEvent{}, ErrInvalidResponse
	}
	amount, err := parseMinor(rawString(payload.PriceAmount))
	if err != nil {
		return providers.GatewayEvent{}, err
	}
	paymentID := rawString(payload.PaymentID)
	return providers.GatewayEvent{GatewayCode: "nowpayments", EventID: paymentID + ":" + payload.Status, EventType: "nowpayments.ipn", ProviderPaymentID: paymentID, ProviderPaymentTx: payload.TxID, Status: MapStatus(payload.Status), AmountMinor: amount, Currency: strings.ToUpper(payload.PriceCurrency), OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) CreateRefund(context.Context, providers.RefundRequest) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, ErrUnsupportedRefund
}

func (p *Provider) QueryRefund(context.Context, string) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, ErrUnsupportedRefund
}

func (p *Provider) readyForAPI() bool {
	return p != nil && p.config.Enabled && p.config.BaseURL != "" && p.config.APIKey != ""
}

func (p *Provider) readyForWebhook() bool {
	return p.readyForAPI() && p.config.IPNSecret != ""
}

func (p *Provider) request(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	request, err := http.NewRequestWithContext(ctx, method, p.config.BaseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", p.config.APIKey)
	response, err := p.client.Do(request)
	if err != nil {
		return nil, 0, &providers.RequestError{GatewayCode: "nowpayments", Retryable: true, ProviderMessage: "network request failed"}
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20+1))
	if err != nil || len(data) > 1<<20 {
		return nil, response.StatusCode, ErrInvalidResponse
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, response.StatusCode, parseRequestError(response.StatusCode, data)
	}
	return data, response.StatusCode, nil
}

func parseRequestError(status int, body []byte) error {
	var payload struct {
		Code    json.RawMessage `json:"code"`
		Message string          `json:"message"`
		Error   string          `json:"error"`
	}
	providerCode, providerMessage := "", ""
	if json.Unmarshal(body, &payload) == nil {
		providerCode = strings.Trim(strings.TrimSpace(string(payload.Code)), `"`)
		providerMessage = strings.TrimSpace(payload.Message)
		if providerMessage == "" {
			providerMessage = strings.TrimSpace(payload.Error)
		}
	}
	return &providers.RequestError{
		GatewayCode: "nowpayments", StatusCode: status, ProviderCode: providerCode,
		ProviderMessage: providerMessage, Retryable: status == http.StatusTooManyRequests || status >= 500,
	}
}

type paymentResponse struct {
	ID            string          `json:"id"`
	PaymentID     json.RawMessage `json:"payment_id"`
	InvoiceURL    string          `json:"invoice_url"`
	Status        string          `json:"payment_status"`
	PriceAmount   string          `json:"price_amount"`
	PriceCurrency string          `json:"price_currency"`
	TxID          string          `json:"pay_address"`
}

func (r *paymentResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID            json.RawMessage `json:"id"`
		PaymentID     json.RawMessage `json:"payment_id"`
		InvoiceURL    string          `json:"invoice_url"`
		Status        string          `json:"payment_status"`
		PriceAmount   json.RawMessage `json:"price_amount"`
		PriceCurrency string          `json:"price_currency"`
		PayAddress    string          `json:"pay_address"`
		TxID          string          `json:"payin_hash"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.ID = rawString(raw.ID)
	if r.ID == "" {
		r.ID = rawString(raw.PaymentID)
	}
	r.InvoiceURL, r.Status, r.PriceCurrency, r.TxID = raw.InvoiceURL, raw.Status, raw.PriceCurrency, raw.TxID
	r.PriceAmount = rawString(raw.PriceAmount)
	r.TxID = strings.TrimSpace(r.TxID)
	return nil
}

type ipnPayload struct {
	PaymentID     json.RawMessage `json:"payment_id"`
	OrderID       string          `json:"order_id"`
	Status        string          `json:"payment_status"`
	PriceAmount   json.RawMessage `json:"price_amount"`
	PriceCurrency string          `json:"price_currency"`
	TxID          string          `json:"payin_hash"`
}

func MapStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "finished":
		return "succeeded"
	case "partially_paid":
		return "pending_review"
	case "failed", "refunded", "expired":
		return "failed"
	default:
		return "pending"
	}
}

func verifyIPN(secret, signature string, body []byte) bool {
	if secret == "" || signature == "" {
		return false
	}
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return false
	}
	canonical, err := canonicalJSON(value)
	if err != nil {
		return false
	}
	mac := hmac.New(sha512.New, []byte(secret))
	_, _ = mac.Write(canonical)
	expected, err := hex.DecodeString(signature)
	return err == nil && hmac.Equal(expected, mac.Sum(nil))
}

func signatureHeader(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func canonicalJSON(value any) ([]byte, error) {
	switch item := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(item))
		for key := range item {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var builder strings.Builder
		builder.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				builder.WriteByte(',')
			}
			encoded, _ := json.Marshal(key)
			builder.Write(encoded)
			builder.WriteByte(':')
			child, err := canonicalJSON(item[key])
			if err != nil {
				return nil, err
			}
			builder.Write(child)
		}
		builder.WriteByte('}')
		return []byte(builder.String()), nil
	case []any:
		var builder strings.Builder
		builder.WriteByte('[')
		for index, childValue := range item {
			if index > 0 {
				builder.WriteByte(',')
			}
			child, err := canonicalJSON(childValue)
			if err != nil {
				return nil, err
			}
			builder.Write(child)
		}
		builder.WriteByte(']')
		return []byte(builder.String()), nil
	case json.Number:
		return []byte(item.String()), nil
	default:
		return json.Marshal(value)
	}
}

func rawString(value json.RawMessage) string {
	if len(value) == 0 || string(value) == "null" {
		return ""
	}
	var stringValue string
	if json.Unmarshal(value, &stringValue) == nil {
		return stringValue
	}
	return strings.Trim(string(value), `"`)
}

func parseMinor(value string) (int64, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) > 2 || len(parts) == 0 || parts[0] == "" {
		return 0, ErrInvalidResponse
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, ErrInvalidResponse
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 && strings.TrimRight(fraction[2:], "0") != "" {
		return 0, ErrInvalidResponse
	}
	fraction = (fraction + "00")[:2]
	minor, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, ErrInvalidResponse
	}
	return whole*100 + minor, nil
}

func formatMinor(value int64) string { return fmt.Sprintf("%d.%02d", value/100, value%100) }
func truncate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "FastImg plan"
	}
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
func urlSegment(value string) string { return strings.Trim(strings.ReplaceAll(value, "/", ""), " ") }
