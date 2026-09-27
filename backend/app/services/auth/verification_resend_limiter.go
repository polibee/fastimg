package authservices

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	cachecontract "github.com/goravel/framework/contracts/cache"

	"goravel/app/facades"
	settingsservices "goravel/app/services/settings"
)

const (
	defaultVerificationResendEmailCooldownSeconds = 60
	defaultVerificationResendIPCooldownSeconds    = 10
	defaultVerificationResendDailyEmailLimit      = 5
	defaultVerificationResendDailyIPLimit         = 20
	maxVerificationResendEmailCooldownSeconds     = 24 * 60 * 60
	maxVerificationResendIPCooldownSeconds        = 60 * 60
	maxVerificationResendDailyEmailLimit          = 1000
	maxVerificationResendDailyIPLimit             = 10000
)

const (
	verificationResendLimitEmailCooldown = "email_cooldown"
	verificationResendLimitIPCooldown    = "ip_cooldown"
	verificationResendLimitDailyEmail    = "daily_email_limit"
	verificationResendLimitDailyIP       = "daily_ip_limit"
)

var ErrVerificationResendStoreUnavailable = errors.New("verification resend rate limit store unavailable")

type VerificationResendPolicy struct {
	Enabled              bool
	EmailCooldownSeconds int
	IPCooldownSeconds    int
	DailyEmailLimit      int
	DailyIPLimit         int
}

type VerificationResendDecision struct {
	Allowed    bool
	Reason     string
	RetryAfter time.Duration
}

type verificationResendState struct {
	EmailCooldownUntil time.Time
	IPCooldownUntil    time.Time
	DailyEmailCount    int
	DailyIPCount       int
}

type VerificationResendRateLimitError struct {
	Reason     string
	RetryAfter time.Duration
}

func (e *VerificationResendRateLimitError) Error() string {
	return fmt.Sprintf("verification resend rate limited: %s", e.Reason)
}

func LoadVerificationResendPolicy(settings *settingsservices.SettingService) VerificationResendPolicy {
	if settings == nil {
		settings = settingsservices.NewSettingService()
	}
	return NormalizeVerificationResendPolicy(VerificationResendPolicy{
		Enabled:              settingBool(settings, "auth.registration.verification_resend_protection_enabled", true),
		EmailCooldownSeconds: settingInt(settings, "auth.registration.verification_resend_email_cooldown_seconds", defaultVerificationResendEmailCooldownSeconds),
		IPCooldownSeconds:    settingInt(settings, "auth.registration.verification_resend_ip_cooldown_seconds", defaultVerificationResendIPCooldownSeconds),
		DailyEmailLimit:      settingInt(settings, "auth.registration.verification_resend_daily_email_limit", defaultVerificationResendDailyEmailLimit),
		DailyIPLimit:         settingInt(settings, "auth.registration.verification_resend_daily_ip_limit", defaultVerificationResendDailyIPLimit),
	})
}

func NormalizeVerificationResendPolicy(policy VerificationResendPolicy) VerificationResendPolicy {
	policy.EmailCooldownSeconds = boundedSetting(policy.EmailCooldownSeconds, defaultVerificationResendEmailCooldownSeconds, 1, maxVerificationResendEmailCooldownSeconds)
	policy.IPCooldownSeconds = boundedSetting(policy.IPCooldownSeconds, defaultVerificationResendIPCooldownSeconds, 1, maxVerificationResendIPCooldownSeconds)
	policy.DailyEmailLimit = boundedSetting(policy.DailyEmailLimit, defaultVerificationResendDailyEmailLimit, 1, maxVerificationResendDailyEmailLimit)
	policy.DailyIPLimit = boundedSetting(policy.DailyIPLimit, defaultVerificationResendDailyIPLimit, 1, maxVerificationResendDailyIPLimit)
	return policy
}

func boundedSetting(value, fallback, min, max int) int {
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	if value < min {
		return min
	}
	return value
}

func verificationResendDecision(policy VerificationResendPolicy, state verificationResendState, now time.Time) VerificationResendDecision {
	if !policy.Enabled {
		return VerificationResendDecision{Allowed: true}
	}
	if state.EmailCooldownUntil.After(now) {
		return VerificationResendDecision{Reason: verificationResendLimitEmailCooldown, RetryAfter: state.EmailCooldownUntil.Sub(now)}
	}
	if state.IPCooldownUntil.After(now) {
		return VerificationResendDecision{Reason: verificationResendLimitIPCooldown, RetryAfter: state.IPCooldownUntil.Sub(now)}
	}
	if state.DailyEmailCount >= policy.DailyEmailLimit {
		return VerificationResendDecision{Reason: verificationResendLimitDailyEmail, RetryAfter: nextDailyWindow(now)}
	}
	if state.DailyIPCount >= policy.DailyIPLimit {
		return VerificationResendDecision{Reason: verificationResendLimitDailyIP, RetryAfter: nextDailyWindow(now)}
	}
	return VerificationResendDecision{Allowed: true}
}

type VerificationResendLimiter struct {
	cache cachecontract.Driver
	now   func() time.Time
}

func NewVerificationResendLimiter() *VerificationResendLimiter {
	return &VerificationResendLimiter{cache: facades.Cache(), now: func() time.Time { return time.Now().UTC() }}
}

func (l *VerificationResendLimiter) Allow(email, ip string, policy VerificationResendPolicy) error {
	policy = NormalizeVerificationResendPolicy(policy)
	if !policy.Enabled {
		return nil
	}
	if l == nil || l.cache == nil {
		return ErrVerificationResendStoreUnavailable
	}
	now := l.now().UTC()
	emailHash := verificationResendHash(strings.ToLower(strings.TrimSpace(email)))
	ipHash := verificationResendHash(strings.TrimSpace(ip))
	emailCooldownKey := "auth:verification-resend:email-cooldown:" + emailHash
	ipCooldownKey := "auth:verification-resend:ip-cooldown:" + ipHash
	emailDailyKey := "auth:verification-resend:email-daily:" + emailHash + ":" + now.Format("20060102")
	ipDailyKey := "auth:verification-resend:ip-daily:" + ipHash + ":" + now.Format("20060102")

	state := verificationResendState{
		DailyEmailCount: int(l.cache.GetInt64(emailDailyKey, 0)),
		DailyIPCount:    int(l.cache.GetInt64(ipDailyKey, 0)),
	}
	if l.cache.Has(emailCooldownKey) {
		state.EmailCooldownUntil = now.Add(time.Duration(policy.EmailCooldownSeconds) * time.Second)
	}
	if l.cache.Has(ipCooldownKey) {
		state.IPCooldownUntil = now.Add(time.Duration(policy.IPCooldownSeconds) * time.Second)
	}
	decision := verificationResendDecision(policy, state, now)
	if !decision.Allowed {
		return &VerificationResendRateLimitError{Reason: decision.Reason, RetryAfter: decision.RetryAfter}
	}

	emailCooldown := time.Duration(policy.EmailCooldownSeconds) * time.Second
	ipCooldown := time.Duration(policy.IPCooldownSeconds) * time.Second
	if !l.cache.Add(emailCooldownKey, int64(1), emailCooldown) {
		if l.cache.Has(emailCooldownKey) {
			return &VerificationResendRateLimitError{Reason: verificationResendLimitEmailCooldown, RetryAfter: emailCooldown}
		}
		return ErrVerificationResendStoreUnavailable
	}
	if !l.cache.Add(ipCooldownKey, int64(1), ipCooldown) {
		l.cache.Forget(emailCooldownKey)
		if l.cache.Has(ipCooldownKey) {
			return &VerificationResendRateLimitError{Reason: verificationResendLimitIPCooldown, RetryAfter: ipCooldown}
		}
		return ErrVerificationResendStoreUnavailable
	}

	dailyTTL := nextDailyWindow(now)
	l.cache.Add(emailDailyKey, int64(0), dailyTTL)
	l.cache.Add(ipDailyKey, int64(0), dailyTTL)
	emailCount, err := l.cache.Increment(emailDailyKey, 1)
	if err != nil {
		l.cache.Forget(emailCooldownKey)
		l.cache.Forget(ipCooldownKey)
		return ErrVerificationResendStoreUnavailable
	}
	ipCount, err := l.cache.Increment(ipDailyKey, 1)
	if err != nil {
		l.cache.Forget(emailCooldownKey)
		l.cache.Forget(ipCooldownKey)
		return ErrVerificationResendStoreUnavailable
	}
	if emailCount > int64(policy.DailyEmailLimit) {
		return &VerificationResendRateLimitError{Reason: verificationResendLimitDailyEmail, RetryAfter: dailyTTL}
	}
	if ipCount > int64(policy.DailyIPLimit) {
		return &VerificationResendRateLimitError{Reason: verificationResendLimitDailyIP, RetryAfter: dailyTTL}
	}
	return nil
}

func verificationResendHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func nextDailyWindow(now time.Time) time.Duration {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day()+1, 0, 0, 0, 0, time.UTC)
	if !next.After(utc) {
		return 24 * time.Hour
	}
	return next.Sub(utc)
}
