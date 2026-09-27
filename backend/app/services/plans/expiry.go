package planservices

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

var errUnsupportedSubscriptionPeriod = errors.New("unsupported subscription period")

const (
	subscriptionStateActive  = "active"
	subscriptionStateExpired = "expired"
)

func normalizeExpiryReminderDays(raw string) []int {
	seen := map[int]struct{}{}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\r' || r == '\t' }) {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value <= 0 || value > 365 {
			continue
		}
		seen[value] = struct{}{}
	}
	days := make([]int, 0, len(seen))
	for value := range seen {
		days = append(days, value)
	}
	sort.Ints(days)
	return days
}

func classifySubscriptionExpiry(endsAt *time.Time, now time.Time) string {
	if endsAt == nil || endsAt.After(now) {
		return subscriptionStateActive
	}
	return subscriptionStateExpired
}

func gracePeriodEndsAt(expiresAt time.Time, days int) time.Time {
	if days < 0 {
		days = 0
	}
	return expiresAt.Add(time.Duration(days) * 24 * time.Hour)
}

func expiryNotificationKey(subscriptionID uint, kind string) string {
	return fmt.Sprintf("subscription:%d:expiry:%s", subscriptionID, strings.TrimSpace(kind))
}

func subscriptionTermEndsAt(startsAt time.Time, billingPeriod string, trialDays int) (time.Time, error) {
	if trialDays < 0 {
		trialDays = 0
	}
	var endsAt time.Time
	switch strings.ToLower(strings.TrimSpace(billingPeriod)) {
	case "monthly":
		endsAt = startsAt.AddDate(0, 1, 0)
	case "yearly":
		endsAt = startsAt.AddDate(1, 0, 0)
	default:
		return time.Time{}, errUnsupportedSubscriptionPeriod
	}
	return endsAt.AddDate(0, 0, trialDays), nil
}

// SubscriptionTermEndsAt converts the immutable billing terms captured on a
// paid order into the subscription expiry timestamp.
func SubscriptionTermEndsAt(startsAt time.Time, billingPeriod string, trialDays int) (time.Time, error) {
	return subscriptionTermEndsAt(startsAt, billingPeriod, trialDays)
}
