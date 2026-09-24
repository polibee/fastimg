package xcash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func Sign(key, nonce, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(nonce))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(key string, headers map[string]string, body []byte, now time.Time, window time.Duration) bool {
	if strings.TrimSpace(key) == "" || window <= 0 {
		return false
	}
	nonce := header(headers, "XC-Nonce")
	timestamp := header(headers, "XC-Timestamp")
	provided := header(headers, "XC-Signature")
	if nonce == "" || timestamp == "" || provided == "" {
		return false
	}
	parsed, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || abs(now.Unix()-parsed) > int64(window/time.Second) {
		return false
	}
	expected, err := hex.DecodeString(provided)
	if err != nil || len(expected) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(nonce))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write(body)
	return hmac.Equal(expected, mac.Sum(nil))
}

func header(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
