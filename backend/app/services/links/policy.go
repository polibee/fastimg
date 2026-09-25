package links

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	HotlinkModeOff     = "off"
	HotlinkModeReferer = "referer"
	HotlinkModeSigned  = "signed"
	HotlinkModeHybrid  = "hybrid"
)

var ErrInvalidHotlinkDomain = errors.New("invalid hotlink domain")

func normalizeHotlinkDomain(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidHotlinkDomain
	}
	parseValue := raw
	if !strings.Contains(raw, "://") {
		parseValue = "https://" + raw
	}
	parsed, err := url.Parse(parseValue)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrInvalidHotlinkDomain
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if host == "" {
		return "", ErrInvalidHotlinkDomain
	}
	if port := parsed.Port(); port != "" {
		host = net.JoinHostPort(host, port)
	}
	return host, nil
}

func normalizeRefererHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	normalized, err := normalizeHotlinkDomain(parsed.Host)
	if err != nil {
		return ""
	}
	return normalized
}

func signMediaURL(mediaID uint, variant string, expiresAt time.Time, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(mediaURLPayload(mediaID, variant, expiresAt)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Stable media URLs deliberately sign an expiry of zero. They are bearer URLs
// intended for forum embeds: revocation is controlled by deleting the media,
// disabling its hotlink policy, or rotating APP_KEY.
func signStableMediaURL(mediaID uint, variant, secret string) string {
	return signMediaURL(mediaID, variant, time.Unix(0, 0).UTC(), secret)
}

func verifyStableMediaURL(mediaID uint, variant, signature, secret string) bool {
	if strings.TrimSpace(secret) == "" || strings.TrimSpace(signature) == "" {
		return false
	}
	expected := signStableMediaURL(mediaID, variant, secret)
	provided, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	want, err := base64.RawURLEncoding.DecodeString(expected)
	if err != nil {
		return false
	}
	return hmac.Equal(provided, want)
}

func verifyMediaURL(mediaID uint, variant string, expiresAt time.Time, signature, secret string, now time.Time) bool {
	if strings.TrimSpace(secret) == "" || strings.TrimSpace(signature) == "" || !expiresAt.After(now) {
		return false
	}
	expected := signMediaURL(mediaID, variant, expiresAt, secret)
	provided, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	want, err := base64.RawURLEncoding.DecodeString(expected)
	if err != nil {
		return false
	}
	return hmac.Equal(provided, want)
}

func mediaURLPayload(mediaID uint, variant string, expiresAt time.Time) string {
	return fmt.Sprintf("%d\n%s\n%d", mediaID, variant, expiresAt.Unix())
}

func allowsHotlink(mode, referer string, allowNoReferer bool, domains []string, signatureValid bool) bool {
	switch mode {
	case HotlinkModeOff:
		return true
	case HotlinkModeSigned:
		return signatureValid
	case HotlinkModeReferer:
		return refererAllowed(referer, allowNoReferer, domains)
	case HotlinkModeHybrid:
		return signatureValid || refererAllowed(referer, allowNoReferer, domains)
	default:
		return false
	}
}

func refererAllowed(referer string, allowNoReferer bool, domains []string) bool {
	host := normalizeRefererHost(referer)
	if host == "" {
		return allowNoReferer
	}
	for _, domain := range domains {
		normalized, err := normalizeHotlinkDomain(domain)
		if err == nil && normalized == host {
			return true
		}
	}
	return false
}
