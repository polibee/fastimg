package media

import "time"

const (
	ExpiryStateActive   = "active"
	ExpiryStateExpiring = "expiring"
	ExpiryStateExpired  = "expired"
)

func ExpiryState(expiresAt, now time.Time, warning time.Duration) string {
	if expiresAt.IsZero() || expiresAt.After(now.Add(warning)) {
		return ExpiryStateActive
	}
	if expiresAt.After(now) {
		return ExpiryStateExpiring
	}
	return ExpiryStateExpired
}
