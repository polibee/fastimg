package developer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
)

const (
	personalTokenRateLimit = 60
	personalIPRateLimit    = 300
	personalRateWindow     = time.Minute
)

var ErrRateLimitStoreUnavailable = errors.New("personal api rate limit store unavailable")

type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter int
	ResetAt    time.Time
}

type rateLimitRecord struct {
	KeyHash         string    `gorm:"column:key_hash"`
	Requests        int       `gorm:"column:requests"`
	WindowStartedAt time.Time `gorm:"column:window_started_at"`
}

// rateLimitDecision applies one request to a fixed window. It is kept pure so
// the boundary behavior can be tested without requiring a database.
func rateLimitDecision(attempts, limit int, startedAt, now time.Time, window time.Duration) (bool, int, int, time.Time) {
	if limit <= 0 || window <= 0 {
		return false, 0, 1, now.Add(window)
	}
	if now.Sub(startedAt) >= window || now.Before(startedAt) {
		attempts = 0
		startedAt = now
	}
	resetAt := startedAt.Add(window)
	next := attempts + 1
	if next > limit {
		return false, 0, retryAfterSeconds(resetAt, now), resetAt
	}
	return true, limit - next, 0, resetAt
}

func retryAfterSeconds(resetAt, now time.Time) int {
	seconds := math.Ceil(resetAt.Sub(now).Seconds())
	if seconds < 1 {
		return 1
	}
	return int(seconds)
}

type rateLimitSubject struct {
	key   string
	limit int
}

type APIRateLimiter struct{}

func NewAPIRateLimiter() *APIRateLimiter { return &APIRateLimiter{} }

func (r *APIRateLimiter) Allow(_ context.Context, tokenID uint, ip string) (RateLimitResult, error) {
	if tokenID == 0 {
		return RateLimitResult{Allowed: false, Limit: personalTokenRateLimit, Remaining: 0, RetryAfter: 1}, nil
	}
	subjects := []rateLimitSubject{{key: "token:" + strconv.FormatUint(uint64(tokenID), 10), limit: personalTokenRateLimit}}
	if trimmedIP := strings.TrimSpace(ip); trimmedIP != "" {
		subjects = append(subjects, rateLimitSubject{key: "ip:" + trimmedIP, limit: personalIPRateLimit})
	}

	now := time.Now().UTC()
	result := RateLimitResult{Allowed: true, Limit: personalTokenRateLimit, Remaining: personalTokenRateLimit, ResetAt: now.Add(personalRateWindow)}
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		for _, subject := range subjects {
			current, err := r.consumeSubject(tx, subject, now)
			if err != nil {
				return err
			}
			if current.Remaining < result.Remaining {
				result.Remaining = current.Remaining
			}
			if current.Limit == personalTokenRateLimit || !current.Allowed {
				result.Limit = current.Limit
			}
			if current.ResetAt.After(result.ResetAt) || !current.Allowed {
				result.ResetAt = current.ResetAt
			}
			if !current.Allowed {
				result.Allowed = false
				result.RetryAfter = current.RetryAfter
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return RateLimitResult{}, errors.Join(ErrRateLimitStoreUnavailable, err)
	}
	return result, nil
}

func (r *APIRateLimiter) consumeSubject(tx orm.Query, subject rateLimitSubject, now time.Time) (RateLimitResult, error) {
	hash := rateLimitKeyHash(subject.key)
	var row rateLimitRecord
	existsQuery := tx.Table("api_token_rate_limits").Where("key_hash = ?", hash)
	exists, err := existsQuery.Exists()
	if err != nil {
		return RateLimitResult{}, err
	}
	if !exists {
		if createErr := tx.Table("api_token_rate_limits").Create(&map[string]any{
			"key_hash": hash, "requests": 1, "window_started_at": now,
			"created_at": now, "updated_at": now,
		}); createErr != nil {
			return RateLimitResult{}, createErr
		}
		return RateLimitResult{Allowed: true, Limit: subject.limit, Remaining: subject.limit - 1, ResetAt: now.Add(personalRateWindow)}, nil
	}
	if err := tx.Table("api_token_rate_limits").Where("key_hash = ?", hash).LockForUpdate().First(&row); err != nil {
		return RateLimitResult{}, err
	}

	allowed, remaining, retryAfter, resetAt := rateLimitDecision(row.Requests, subject.limit, row.WindowStartedAt, now, personalRateWindow)
	if !allowed {
		return RateLimitResult{Allowed: false, Limit: subject.limit, Remaining: remaining, RetryAfter: retryAfter, ResetAt: resetAt}, nil
	}
	startedAt := row.WindowStartedAt
	requests := row.Requests + 1
	if now.Sub(startedAt) >= personalRateWindow || now.Before(startedAt) {
		startedAt = now
		requests = 1
	}
	if _, err := tx.Table("api_token_rate_limits").Where("key_hash = ?", hash).Update(map[string]any{
		"requests": requests, "window_started_at": startedAt, "updated_at": now,
	}); err != nil {
		return RateLimitResult{}, err
	}
	return RateLimitResult{Allowed: true, Limit: subject.limit, Remaining: remaining, ResetAt: resetAt}, nil
}

func rateLimitKeyHash(key string) string {
	digest := sha256.Sum256([]byte(key))
	return hex.EncodeToString(digest[:])
}
