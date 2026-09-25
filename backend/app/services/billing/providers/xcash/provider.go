package xcash

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goravel/app/services/billing/providers"
)

var ErrInvalidResponse = errors.New("invalid xcash response")
var ErrUnsupportedRefund = errors.New("xcash refunds require manual processing")

type Config struct {
	Enabled     bool
	BaseURL     string
	AppID       string
	HMACKey     string
	Timeout     time.Duration
	CallbackURL string
	ReturnURL   string
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
	if !p.ready() || request.OrderNo == "" || request.AmountMinor <= 0 {
		return providers.PaymentSession{}, providers.ErrGatewayUnavailable
	}
	payload := map[string]any{
		"out_no": request.OrderNo, "title": truncateTitle(request.Description), "currency": request.Currency,
		"amount": formatMinor(request.AmountMinor), "duration": 30,
		"notify_url": p.config.CallbackURL, "return_url": orderReturnURL(p.config.ReturnURL, request.OrderNo),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return providers.PaymentSession{}, err
	}
	responseBody, _, err := p.request(ctx, http.MethodPost, "/v1/invoice", body, true)
	if err != nil {
		return providers.PaymentSession{}, err
	}
	var response invoiceResponse
	if err := json.Unmarshal(responseBody, &response); err != nil || response.SysNo == "" || response.PayURL == "" {
		return providers.PaymentSession{}, ErrInvalidResponse
	}
	return providers.PaymentSession{GatewayCode: "xcash", ProviderPaymentID: response.SysNo, ProviderOrderID: response.OutNo, Status: MapStatus(response.Status, response.Confirmed, response.RiskLevel), CheckoutURL: response.PayURL, OccurredAt: time.Now().Unix()}, nil
}

func orderReturnURL(template, orderNo string) string {
	return strings.ReplaceAll(template, "{order_id}", orderNo)
}

func (p *Provider) QueryPayment(ctx context.Context, request providers.QueryPaymentRequest) (providers.GatewayPayment, error) {
	if !p.ready() || request.ProviderPaymentID == "" {
		return providers.GatewayPayment{}, providers.ErrGatewayUnavailable
	}
	responseBody, _, err := p.request(ctx, http.MethodGet, "/v1/invoice/"+urlPath(request.ProviderPaymentID), nil, false)
	if err != nil {
		return providers.GatewayPayment{}, err
	}
	var response invoiceResponse
	if err := json.Unmarshal(responseBody, &response); err != nil || response.SysNo == "" {
		return providers.GatewayPayment{}, ErrInvalidResponse
	}
	amount, err := ParseMinor(response.Amount)
	if err != nil {
		return providers.GatewayPayment{}, err
	}
	return providers.GatewayPayment{GatewayCode: "xcash", ProviderPaymentID: response.SysNo, ProviderPaymentTx: response.Payment.Hash, Status: MapStatus(response.Status, response.Confirmed, response.RiskLevel), AmountMinor: amount, Currency: response.Currency, OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) VerifyWebhook(ctx context.Context, request providers.WebhookRequest) (providers.GatewayEvent, error) {
	if !p.ready() || header(request.Headers, "XC-Appid") != p.config.AppID || !VerifySignature(p.config.HMACKey, request.Headers, request.Body, time.Now(), 5*time.Minute) {
		return providers.GatewayEvent{}, providers.ErrGatewayUnavailable
	}
	var payload webhookPayload
	if err := json.Unmarshal(request.Body, &payload); err != nil || payload.Type != "invoice" || payload.Data.SysNo == "" {
		return providers.GatewayEvent{}, ErrInvalidResponse
	}
	data := payload.Data
	amount, err := ParseMinor(data.Amount)
	if err != nil || data.Currency == "" {
		queried, queryErr := p.QueryPayment(ctx, providers.QueryPaymentRequest{ProviderPaymentID: data.SysNo})
		if queryErr != nil {
			return providers.GatewayEvent{}, ErrInvalidResponse
		}
		amount, data.Currency, data.Status = queried.AmountMinor, queried.Currency, queried.Status
		if data.Hash == "" {
			data.Hash = queried.ProviderPaymentTx
		}
	}
	eventID := data.SysNo
	if data.Hash != "" {
		eventID += ":" + data.Hash
	}
	return providers.GatewayEvent{GatewayCode: "xcash", EventID: eventID, EventType: "xcash." + payload.Type, ProviderPaymentID: data.SysNo, ProviderPaymentTx: data.Hash, Status: MapStatus(data.Status, data.Confirmed, data.RiskLevel), AmountMinor: amount, Currency: data.Currency, OccurredAt: time.Now().Unix()}, nil
}

func (p *Provider) CreateRefund(context.Context, providers.RefundRequest) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, ErrUnsupportedRefund
}

func (p *Provider) QueryRefund(context.Context, string) (providers.GatewayRefund, error) {
	return providers.GatewayRefund{}, ErrUnsupportedRefund
}

func (p *Provider) ready() bool {
	return p != nil && p.config.Enabled && p.config.BaseURL != "" && p.config.AppID != "" && p.config.HMACKey != ""
}

func (p *Provider) request(ctx context.Context, method, path string, body []byte, signed bool) ([]byte, int, error) {
	request, err := http.NewRequestWithContext(ctx, method, p.config.BaseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Content-Type", "application/json")
	if signed {
		nonceBytes := make([]byte, 16)
		if _, err := rand.Read(nonceBytes); err != nil {
			return nil, 0, err
		}
		nonce := hex.EncodeToString(nonceBytes)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		request.Header.Set("XC-Appid", p.config.AppID)
		request.Header.Set("XC-Timestamp", timestamp)
		request.Header.Set("XC-Nonce", nonce)
		request.Header.Set("XC-Signature", Sign(p.config.HMACKey, nonce, timestamp, body))
	}
	response, err := p.client.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("xcash request failed: %w", err)
	}
	defer response.Body.Close()
	limited, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20+1))
	if readErr != nil {
		return nil, response.StatusCode, readErr
	}
	if len(limited) > 1<<20 {
		return nil, response.StatusCode, ErrInvalidResponse
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, response.StatusCode, fmt.Errorf("xcash request returned status %d", response.StatusCode)
	}
	return limited, response.StatusCode, nil
}

type invoiceResponse struct {
	SysNo     string `json:"sys_no"`
	OutNo     string `json:"out_no"`
	PayURL    string `json:"pay_url"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	Confirmed bool   `json:"confirmed"`
	RiskLevel string `json:"risk_level"`
	ExpiresAt string `json:"expires_at"`
	Payment   struct {
		Hash string `json:"hash"`
	} `json:"payment"`
}

type webhookPayload struct {
	Type string `json:"type"`
	Data struct {
		SysNo     string `json:"sys_no"`
		Currency  string `json:"currency"`
		Amount    string `json:"amount"`
		Hash      string `json:"hash"`
		Status    string `json:"status"`
		Confirmed bool   `json:"confirmed"`
		RiskLevel string `json:"risk_level"`
	} `json:"data"`
}

func MapStatus(status string, confirmed bool, riskLevel string) string {
	if strings.EqualFold(riskLevel, "high") || strings.EqualFold(riskLevel, "critical") {
		return "pending_review"
	}
	if confirmed {
		return "succeeded"
	}
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "confirmed":
		return "succeeded"
	case "expired", "failed", "wrong_network", "canceled", "cancelled":
		return "failed"
	case "underpaid", "overpaid", "pending_review":
		return "pending_review"
	default:
		return "pending"
	}
}

func ParseMinor(value string) (int64, error) {
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
	minorFraction, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, ErrInvalidResponse
	}
	if whole > (int64(^uint64(0)>>1)-minorFraction)/100 {
		return 0, ErrInvalidResponse
	}
	return whole*100 + minorFraction, nil
}

func formatMinor(value int64) string { return fmt.Sprintf("%d.%02d", value/100, value%100) }

func truncateTitle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "FastImg plan"
	}
	if len(value) > 32 {
		return value[:32]
	}
	return value
}

func urlPath(value string) string { return strings.Trim(strings.ReplaceAll(value, "/", ""), " ") }
