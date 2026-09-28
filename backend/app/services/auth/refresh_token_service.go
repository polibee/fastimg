package authservices

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	RefreshTokenCookieName = "go_vue_admin_refresh"
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
)

var ErrRefreshTokenInvalid = errors.New("invalid refresh token")
var ErrRefreshTokenStoreUnavailable = errors.New("refresh token store unavailable")

type refreshTokenStore interface {
	Put(key string, value any, ttl time.Duration) error
	GetString(key string, def ...string) string
	Forget(key string) bool
	Probe() error
	RevokeAll(userID uint) error
}

type RefreshTokenService struct {
	store refreshTokenStore
}

func NewRefreshTokenService(store refreshTokenStore) *RefreshTokenService {
	return &RefreshTokenService{store: store}
}

func (s *RefreshTokenService) Issue(userID uint) (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(buffer)
	if err := s.store.Put(refreshTokenKey(raw), strconv.FormatUint(uint64(userID), 10), RefreshTokenTTL()); err != nil {
		return "", fmt.Errorf("%w: %v", ErrRefreshTokenStoreUnavailable, err)
	}
	return raw, nil
}

// RefreshTokenTTL is shared by the durable store and the browser cookie. A
// non-positive or malformed value falls back to the safe 30-day default;
// durable sessions are never made immortal by accident.
func RefreshTokenTTL() time.Duration {
	minutes, err := strconv.Atoi(strings.TrimSpace(os.Getenv("JWT_REFRESH_TTL")))
	if err != nil || minutes <= 0 {
		return defaultRefreshTokenTTL
	}
	return time.Duration(minutes) * time.Minute
}

func RefreshTokenMaxAge() int {
	return int(RefreshTokenTTL() / time.Second)
}

func (s *RefreshTokenService) Consume(raw string) (uint, error) {
	if err := s.probe(); err != nil {
		return 0, err
	}
	if raw == "" {
		return 0, ErrRefreshTokenInvalid
	}
	key := refreshTokenKey(raw)
	value := s.store.GetString(key)
	if value == "" {
		return 0, ErrRefreshTokenInvalid
	}
	s.store.Forget(key)
	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil || userID == 0 {
		return 0, ErrRefreshTokenInvalid
	}
	return uint(userID), nil
}

func (s *RefreshTokenService) probe() error {
	if err := s.store.Probe(); err != nil {
		return fmt.Errorf("%w: %v", ErrRefreshTokenStoreUnavailable, err)
	}
	return nil
}

func (s *RefreshTokenService) Revoke(raw string) {
	if raw != "" {
		s.store.Forget(refreshTokenKey(raw))
	}
}

func (s *RefreshTokenService) RevokeAll(userID uint) error {
	if err := s.probe(); err != nil {
		return err
	}
	return s.store.RevokeAll(userID)
}

func refreshTokenKey(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return "auth:refresh:" + base64.RawURLEncoding.EncodeToString(hash[:])
}
