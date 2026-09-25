package settingsservices

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"goravel/app/facades"
)

const secretPlaceholder = "__configured__"
const encryptedSecretPrefix = "enc:v1:"

var ErrSecretKeyUnavailable = errors.New("application encryption key is not configured")

// IsSecretKey keeps provider credentials out of API responses and makes the
// setting boundary explicit. Bearer API tokens use one-way hashes elsewhere;
// provider credentials are reversible because the payment adapter must send
// them to the provider.
func IsSecretKey(key string) bool {
	switch strings.TrimSpace(key) {
	case "payment.xcash.app_id", "payment.xcash.hmac_key",
		"payment.nowpayments.api_key", "payment.nowpayments.ipn_secret",
		"payment.paypal.client_id", "payment.paypal.client_secret", "payment.paypal.webhook_id":
		return true
	default:
		return false
	}
}

func SecretPlaceholder() string { return secretPlaceholder }

func encryptSecret(value string) (string, error) {
	key := applicationKey()
	if key == nil {
		return "", ErrSecretKeyUnavailable
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	encoded := base64.RawStdEncoding.EncodeToString(append(nonce, ciphertext...))
	return encryptedSecretPrefix + encoded, nil
}

func decryptSecret(value string) (string, error) {
	if !strings.HasPrefix(value, encryptedSecretPrefix) {
		return value, nil
	}
	key := applicationKey()
	if key == nil {
		return "", ErrSecretKeyUnavailable
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, encryptedSecretPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted setting")
	}
	plaintext, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func applicationKey() []byte {
	value := strings.TrimSpace(facades.Config().GetString("app.key", ""))
	if value == "" {
		return nil
	}
	hash := sha256.Sum256([]byte(value))
	return hash[:]
}
