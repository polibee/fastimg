package emailservices

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	frameworkmail "github.com/goravel/framework/mail"

	settingsservices "goravel/app/services/settings"
)

const (
	ProviderSMTP   = "smtp"
	ProviderAliyun = "aliyun"
	ProviderResend = "resend"
)

var (
	ErrEmailDisabled         = errors.New("email service is disabled")
	ErrEmailNotConfigured    = errors.New("email service is not configured")
	ErrEmailProviderRejected = errors.New("email provider rejected the request")
)

// Message is the provider-neutral transactional email contract. Providers are
// intentionally limited to what registration verification needs today so a
// provider cannot leak transport-specific fields into the auth domain.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

type Service struct {
	settings *settingsservices.SettingService
	resolve  func(key, fallback string) string
	client   *http.Client
	now      func() time.Time
}

func NewService() *Service {
	settings := settingsservices.NewSettingService()
	return &Service{
		settings: settings,
		resolve:  settings.Resolve,
		client:   &http.Client{Timeout: 10 * time.Second},
		now:      time.Now,
	}
}

func (s *Service) Send(ctx context.Context, message Message) error {
	if strings.TrimSpace(message.To) == "" || strings.TrimSpace(message.Subject) == "" {
		return ErrEmailNotConfigured
	}
	if !settingBool(s.resolveSetting, "email.enabled", false) {
		return ErrEmailDisabled
	}
	provider := strings.ToLower(strings.TrimSpace(s.resolveSetting("email.provider", ProviderSMTP)))
	from := strings.TrimSpace(s.resolveSetting("email.from_address", ""))
	fromName := strings.TrimSpace(s.resolveSetting("email.from_name", "FastImg"))
	if from == "" {
		return ErrEmailNotConfigured
	}
	if !validMailbox(from) || !validMailbox(message.To) || strings.ContainsAny(fromName, "\r\n") {
		return ErrEmailNotConfigured
	}

	switch provider {
	case ProviderSMTP:
		return s.sendSMTP(message, from, fromName)
	case ProviderAliyun:
		return s.sendAliyun(ctx, message, from, fromName)
	case ProviderResend:
		return s.sendResend(ctx, message, from, fromName)
	default:
		return ErrEmailNotConfigured
	}
}

func (s *Service) sendSMTP(message Message, from, fromName string) error {
	host := strings.TrimSpace(s.resolveSetting("email.smtp.host", ""))
	if host == "" {
		return ErrEmailNotConfigured
	}
	encryption := strings.ToLower(strings.TrimSpace(s.resolveSetting("email.smtp.encryption", "starttls")))
	port := settingInt(s.resolveSetting, "email.smtp.port", 0)
	if port == 0 {
		if encryption == "ssl" {
			port = 465
		} else if encryption == "none" {
			port = 25
		} else {
			port = 587
		}
	}
	username := strings.TrimSpace(s.resolveSetting("email.smtp.username", ""))
	password := s.resolveSetting("email.smtp.password", "")
	email := frameworkmail.NewEmail()
	email.From = formatAddress(from, fromName)
	email.To = []string{message.To}
	email.Subject = message.Subject
	email.Text = []byte(message.Text)
	email.HTML = []byte(message.HTML)
	if replyTo := strings.TrimSpace(s.resolveSetting("email.reply_to", "")); replyTo != "" {
		if !validMailbox(replyTo) {
			return ErrEmailNotConfigured
		}
		email.ReplyTo = []string{replyTo}
	}

	var smtpAuth = plainAuth(username, password, host)
	address := netJoinHostPort(host, port)
	config := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	switch encryption {
	case "ssl":
		return email.SendWithTLS(address, smtpAuth, config)
	case "none":
		return email.Send(address, smtpAuth)
	case "starttls", "tls", "":
		return email.SendWithStartTLS(address, smtpAuth, config)
	default:
		return ErrEmailNotConfigured
	}
}

func (s *Service) sendResend(ctx context.Context, message Message, from, fromName string) error {
	apiKey := s.resolveSetting("email.resend.api_key", "")
	if strings.TrimSpace(apiKey) == "" {
		return ErrEmailNotConfigured
	}
	endpoint := strings.TrimRight(strings.TrimSpace(s.resolveSetting("email.resend.endpoint", "https://api.resend.com")), "/") + "/emails"
	payload := map[string]any{
		"from":    formatAddress(from, fromName),
		"to":      []string{message.To},
		"subject": message.Subject,
		"text":    message.Text,
		"html":    message.HTML,
	}
	if replyTo := strings.TrimSpace(s.resolveSetting("email.reply_to", "")); replyTo != "" {
		if !validMailbox(replyTo) {
			return ErrEmailNotConfigured
		}
		payload["reply_to"] = []string{replyTo}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return ErrEmailProviderRejected
	}
	return nil
}

func (s *Service) sendAliyun(ctx context.Context, message Message, from, fromName string) error {
	accessKeyID := s.resolveSetting("email.aliyun.access_key_id", "")
	accessKeySecret := s.resolveSetting("email.aliyun.access_key_secret", "")
	accountName := strings.TrimSpace(s.resolveSetting("email.aliyun.account_name", from))
	if strings.TrimSpace(accessKeyID) == "" || strings.TrimSpace(accessKeySecret) == "" || accountName == "" {
		return ErrEmailNotConfigured
	}
	endpoint := strings.TrimRight(strings.TrimSpace(s.resolveSetting("email.aliyun.endpoint", "https://dm.aliyuncs.com")), "/")
	params := map[string]string{
		"Action":           "SingleSendMail",
		"AccessKeyId":      accessKeyID,
		"AccountName":      accountName,
		"AddressType":      "1",
		"Format":           "JSON",
		"HtmlBody":         message.HTML,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   randomNonce(),
		"SignatureVersion": "1.0",
		"Subject":          message.Subject,
		"TextBody":         message.Text,
		"ToAddress":        message.To,
		"Version":          "2015-11-23",
		"ReplyToAddress":   "false",
		"Timestamp":        s.now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	if fromName != "" {
		params["FromAlias"] = fromName
	}
	if replyTo := strings.TrimSpace(s.resolveSetting("email.reply_to", "")); replyTo != "" {
		if !validMailbox(replyTo) {
			return ErrEmailNotConfigured
		}
		params["ReplyAddress"] = replyTo
	}
	params["Signature"] = aliyunSignature(http.MethodPost, params, accessKeySecret)
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var result struct {
		RequestID string `json:"RequestId"`
		Code      string `json:"Code"`
		Message   string `json:"Message"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&result); err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || result.Code != "" {
		return ErrEmailProviderRejected
	}
	return nil
}

func settingBool(resolve func(string, string) string, key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(resolve(key, "")))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func settingInt(resolve func(string, string) string, key string, fallback int) int {
	value := strings.TrimSpace(resolve(key, ""))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 65535 {
		return fallback
	}
	return parsed
}

func (s *Service) resolveSetting(key, fallback string) string {
	if s.resolve != nil {
		return s.resolve(key, fallback)
	}
	if s.settings != nil {
		return s.settings.Resolve(key, fallback)
	}
	return fallback
}

func formatAddress(address, name string) string {
	if strings.TrimSpace(name) == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", name, address)
}

func netJoinHostPort(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func plainAuth(username, password, host string) smtp.Auth {
	if strings.TrimSpace(username) == "" {
		return nil
	}
	return smtp.PlainAuth("", username, password, host)
}

func randomNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func aliyunSignature(method string, params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "Signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, percentEncode(key)+"="+percentEncode(params[key]))
	}
	canonical := strings.Join(parts, "&")
	stringToSign := strings.ToUpper(method) + "&%2F&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func percentEncode(value string) string {
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

func validMailbox(value string) bool {
	if strings.ContainsAny(value, "\r\n") {
		return false
	}
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	return err == nil && parsed.Address == strings.TrimSpace(value)
}
