package controllers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	authservices "goravel/app/services/auth"
	emailservices "goravel/app/services/email"
	settingsservices "goravel/app/services/settings"
)

type registrationRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirmation"`
	TurnstileToken  string `json:"turnstile_token"`
}

func (r *AuthController) RegistrationPolicy(ctx http.Context) http.Response {
	policy := authservices.LoadRegistrationPolicy()
	return ctx.Response().Success().Json(http.Json{"data": http.Json{
		"registration_enabled":        policy.RegistrationEnabled,
		"registration_turnstile":      policy.RegistrationTurnstile,
		"login_turnstile":             policy.LoginTurnstile,
		"email_verification_required": policy.EmailVerificationRequired,
		"turnstile_site_key":          policy.TurnstileSiteKey,
	}})
}

func (r *AuthController) Register(ctx http.Context) http.Response {
	policy := authservices.LoadRegistrationPolicy()
	if !policy.RegistrationEnabled {
		return ctx.Response().Status(403).Json(http.Json{"code": "AUTH_REGISTRATION_DISABLED"})
	}
	var payload registrationRequest
	if err := ctx.Request().Bind(&payload); err != nil || strings.TrimSpace(payload.Password) != strings.TrimSpace(payload.PasswordConfirm) {
		return ctx.Response().Status(422).Json(http.Json{"code": "VALIDATION_ERROR"})
	}
	if response := verifyTurnstilePolicy(ctx, policy.RegistrationTurnstile, payload.TurnstileToken); response != nil {
		return response
	}
	rateLimiter := authservices.NewLoginRateLimiter()
	allowed, err := rateLimiter.Allow(payload.Email, ctx.Request().Ip())
	if err != nil {
		return ctx.Response().Status(503).Json(http.Json{"code": "AUTH_RATE_LIMIT_STORE_UNAVAILABLE"})
	}
	if !allowed {
		return ctx.Response().Status(429).Json(http.Json{"code": "AUTH_RATE_LIMITED"})
	}
	result, err := authservices.NewRegistrationService().Register(authservices.RegistrationInput{
		Name: payload.Name, Email: payload.Email, Password: payload.Password,
	}, policy)
	if err != nil {
		switch {
		case errors.Is(err, authservices.ErrRegistrationInvalid):
			return ctx.Response().Status(422).Json(http.Json{"code": "VALIDATION_ERROR"})
		case errors.Is(err, authservices.ErrRegistrationEmailTaken):
			return ctx.Response().Status(409).Json(http.Json{"code": "AUTH_EMAIL_TAKEN"})
		case errors.Is(err, authservices.ErrRegistrationEmailBlocked):
			return ctx.Response().Status(403).Json(http.Json{"code": "AUTH_EMAIL_NOT_ALLOWED"})
		default:
			return ctx.Response().Status(500).Json(http.Json{"code": "AUTH_REGISTRATION_FAILED"})
		}
	}
	if result.VerificationRequired {
		if err := sendVerificationEmail(result.User.Email, result.VerificationToken); err != nil {
			return ctx.Response().Status(503).Json(http.Json{"code": "AUTH_EMAIL_UNAVAILABLE"})
		}
	}
	recordAudit(result.User.ID, "auth.register", map[string]any{"email_verification_required": result.VerificationRequired})
	return ctx.Response().Status(201).Json(http.Json{"data": http.Json{
		"user": result.User.Public(), "verification_required": result.VerificationRequired,
	}})
}

func (r *AuthController) VerifyEmail(ctx http.Context) http.Response {
	user, err := authservices.NewRegistrationService().Verify(ctx.Request().Query("token"))
	if err != nil {
		return ctx.Response().Status(422).Json(http.Json{"code": "AUTH_EMAIL_VERIFICATION_EXPIRED"})
	}
	recordAudit(user.ID, "auth.email_verified", nil)
	return ctx.Response().Success().Json(http.Json{"data": http.Json{"verified": true}})
}

func (r *AuthController) ResendVerification(ctx http.Context) http.Response {
	policy := authservices.LoadRegistrationPolicy()
	if !policy.EmailVerificationRequired {
		return ctx.Response().Success().Json(http.Json{"data": http.Json{"sent": true}})
	}
	var payload struct {
		Email          string `json:"email"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := ctx.Request().Bind(&payload); err != nil {
		return ctx.Response().Status(422).Json(http.Json{"code": "VALIDATION_ERROR"})
	}
	if response := verifyTurnstilePolicy(ctx, policy.RegistrationTurnstile, payload.TurnstileToken); response != nil {
		return response
	}
	email, err := authservices.NormalizeRegistrationEmail(payload.Email)
	if err != nil {
		return ctx.Response().Status(422).Json(http.Json{"code": "VALIDATION_ERROR"})
	}
	if err := authservices.NewVerificationResendLimiter().Allow(email, ctx.Request().Ip(), authservices.LoadVerificationResendPolicy(nil)); err != nil {
		var rateLimitErr *authservices.VerificationResendRateLimitError
		if errors.As(err, &rateLimitErr) {
			retryAfter := int((rateLimitErr.RetryAfter + time.Second - 1) / time.Second)
			if retryAfter < 1 {
				retryAfter = 1
			}
			return ctx.Response().Status(429).Json(http.Json{
				"code":                "AUTH_VERIFICATION_RESEND_RATE_LIMITED",
				"retry_after_seconds": retryAfter,
			})
		}
		return ctx.Response().Status(503).Json(http.Json{"code": "AUTH_RATE_LIMIT_STORE_UNAVAILABLE"})
	}
	var user models.User
	if err := facades.Orm().Query().Where("email = ?", email).First(&user); err == nil && user.EmailVerifiedAt == nil && user.Status == "active" {
		token, tokenErr := authservices.NewRegistrationService().NewVerificationToken(user.ID, policy.VerificationExpiryMinutes)
		if tokenErr != nil || sendVerificationEmail(user.Email, token) != nil {
			return ctx.Response().Status(503).Json(http.Json{"code": "AUTH_EMAIL_UNAVAILABLE"})
		}
	}
	return ctx.Response().Success().Json(http.Json{"data": http.Json{"sent": true}})
}

func verifyTurnstilePolicy(ctx http.Context, enabled bool, token string) http.Response {
	if !enabled {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return ctx.Response().Status(422).Json(http.Json{"code": "AUTH_CAPTCHA_REQUIRED"})
	}
	secret := settingsservices.NewSettingService().Resolve("auth.turnstile.secret_key", "")
	if strings.TrimSpace(secret) == "" {
		return ctx.Response().Status(503).Json(http.Json{"code": "AUTH_CAPTCHA_UNAVAILABLE"})
	}
	if err := authservices.VerifyTurnstile(ctx.Context(), token, ctx.Request().Ip(), secret); err != nil {
		if errors.Is(err, authservices.ErrTurnstileRejected) {
			return ctx.Response().Status(403).Json(http.Json{"code": "AUTH_CAPTCHA_FAILED"})
		}
		return ctx.Response().Status(503).Json(http.Json{"code": "AUTH_CAPTCHA_UNAVAILABLE"})
	}
	return nil
}

func sendVerificationEmail(email, token string) error {
	baseURL := strings.TrimRight(settingsservices.NewSettingService().Resolve("site_url", facades.Config().GetString("app.url", "http://127.0.0.1:53083")), "/")
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)
	return emailservices.NewService().Send(context.Background(), emailservices.Message{
		To:      email,
		Subject: "Verify your FastImg email",
		Text:    fmt.Sprintf("Open this link to verify your FastImg email: %s\nThis link expires soon and can only be used once.", verificationURL),
		HTML:    fmt.Sprintf("<p>Verify your FastImg email:</p><p><a href=\"%s\">%s</a></p><p>This link expires soon and can only be used once.</p>", verificationURL, verificationURL),
	})
}
