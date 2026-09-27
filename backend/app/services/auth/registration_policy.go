package authservices

import (
	"fmt"
	"net/mail"
	"strings"

	settingsservices "goravel/app/services/settings"
)

const (
	defaultVerificationExpiryMinutes = 30
	minVerificationExpiryMinutes     = 5
	maxVerificationExpiryMinutes     = 1440
)

// RegistrationPolicy is the public-safe authentication policy. It deliberately
// never contains the Turnstile secret or the private allow-list domains.
type RegistrationPolicy struct {
	RegistrationEnabled       bool
	RegistrationTurnstile     bool
	LoginTurnstile            bool
	EmailVerificationRequired bool
	EmailWhitelistEnabled     bool
	TurnstileSiteKey          string
	VerificationExpiryMinutes int
}

func LoadRegistrationPolicy() RegistrationPolicy {
	settings := settingsservices.NewSettingService()
	globalTurnstile := settingBool(settings, "auth.turnstile.enabled", false)
	return RegistrationPolicy{
		RegistrationEnabled:       settingBool(settings, "auth.registration.enabled", true),
		RegistrationTurnstile:     globalTurnstile && settingBool(settings, "auth.registration.turnstile_enabled", false),
		LoginTurnstile:            globalTurnstile && settingBool(settings, "auth.login.turnstile_enabled", false),
		EmailVerificationRequired: settingBool(settings, "auth.registration.email_verification_enabled", false),
		EmailWhitelistEnabled:     settingBool(settings, "auth.registration.email_whitelist_enabled", false),
		TurnstileSiteKey:          strings.TrimSpace(settings.Resolve("auth.turnstile.site_key", "")),
		VerificationExpiryMinutes: NormalizeVerificationExpiry(settingInt(settings, "auth.registration.verification_expiry_minutes", defaultVerificationExpiryMinutes)),
	}
}

func settingBool(settings *settingsservices.SettingService, key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(settings.Resolve(key, "")))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func settingInt(settings *settingsservices.SettingService, key string, fallback int) int {
	value := strings.TrimSpace(settings.Resolve(key, ""))
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}

func NormalizeVerificationExpiry(minutes int) int {
	if minutes == 0 {
		return defaultVerificationExpiryMinutes
	}
	if minutes < minVerificationExpiryMinutes {
		return minVerificationExpiryMinutes
	}
	if minutes > maxVerificationExpiryMinutes {
		return maxVerificationExpiryMinutes
	}
	return minutes
}

func ParseWhitelistDomains(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' ' })
	seen := make(map[string]struct{}, len(parts))
	domains := make([]string, 0, len(parts))
	for _, part := range parts {
		domain := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(part, "*.")))
		domain = strings.TrimSuffix(domain, ".")
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		domains = append(domains, domain)
	}
	return domains
}

func EmailAllowedByWhitelist(email string, domains []string) bool {
	parsed, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return false
	}
	address := strings.ToLower(strings.TrimSpace(parsed.Address))
	at := strings.LastIndexByte(address, '@')
	if at < 1 || at == len(address)-1 {
		return false
	}
	domain := strings.TrimSuffix(address[at+1:], ".")
	for _, allowed := range domains {
		allowed = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(allowed), "."))
		if domain == allowed {
			return true
		}
	}
	return false
}
