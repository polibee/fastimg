package authservices

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const turnstileSiteVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

var (
	ErrTurnstileUnavailable = errors.New("turnstile verification unavailable")
	ErrTurnstileRejected    = errors.New("turnstile token rejected")
)

type turnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// VerifyTurnstile performs the server-side Cloudflare validation. Tokens and
// secrets are never included in errors or logs. Cloudflare documents this as
// a form or JSON POST and tokens are single-use, so callers must not retry a
// rejected token blindly.
func VerifyTurnstile(ctx context.Context, token, remoteIP, secret string) error {
	token = strings.TrimSpace(token)
	secret = strings.TrimSpace(secret)
	if token == "" || len(token) > 2048 || secret == "" {
		return ErrTurnstileUnavailable
	}
	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", strings.TrimSpace(remoteIP))
	}
	requestContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodPost, turnstileSiteVerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return ErrTurnstileUnavailable
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return ErrTurnstileUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ErrTurnstileUnavailable
	}
	var result turnstileResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return ErrTurnstileUnavailable
	}
	if !result.Success {
		return ErrTurnstileRejected
	}
	return nil
}
