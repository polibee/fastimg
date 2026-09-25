package planservices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

var (
	ErrBandwidthQuotaExceeded = errors.New("monthly bandwidth quota exceeded")
	ErrBandwidthUnavailable   = errors.New("bandwidth metering unavailable")
)

// bandwidthUsagePeriod is deliberately UTC so a download is charged to one
// deterministic billing month regardless of the server's local timezone.
func bandwidthUsagePeriod(now time.Time) string { return now.UTC().Format("2006-01") }

func validateBandwidthUsageRecord(record quota.UsageRecord) error {
	if record.UserID == 0 || record.ResourceType != "bandwidth" || record.Delta <= 0 || record.SourceType != "download" || strings.TrimSpace(record.SourceID) == "" || record.PeriodKey == "" {
		return fmt.Errorf("%w: invalid bandwidth usage record", quota.ErrInvalidUsageRecord)
	}
	if _, err := time.Parse("2006-01", record.PeriodKey); err != nil {
		return fmt.Errorf("%w: bandwidth period must be YYYY-MM", quota.ErrInvalidUsageRecord)
	}
	return nil
}

// RecordBandwidthUsage charges the owning member for bytes that are about to
// be returned. It is transactionally serialized with the user's subscription
// and idempotent by the access-log source ID.
func RecordBandwidthUsage(_ context.Context, userID uint, bytes int64, sourceID string, now time.Time) error {
	record := quota.UsageRecord{
		UserID: userID, ResourceType: "bandwidth", Delta: bytes,
		SourceType: "download", SourceID: sourceID, PeriodKey: bandwidthUsagePeriod(now),
	}
	if err := validateBandwidthUsageRecord(record); err != nil {
		return err
	}
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var user models.User
		if err := tx.Where("id = ?", userID).LockForUpdate().First(&user); err != nil {
			return fmt.Errorf("%w: owner lookup: %v", ErrBandwidthUnavailable, err)
		}
		// A repeated delivery/access retry must not be charged twice.
		var existing models.UsageLedger
		if err := tx.Where("user_id = ? AND resource_type = ? AND source_type = ? AND source_id = ? AND period_key = ?", userID, "bandwidth", "download", sourceID, record.PeriodKey).First(&existing); err == nil && existing.ID > 0 {
			return nil
		}
		subscription, err := EnsureActiveSubscriptionWithQuery(tx, userID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrBandwidthUnavailable, err)
		}
		entitlement, err := ParseSubscriptionSnapshot(subscription.EntitlementSnapshotJSON)
		if err != nil {
			return fmt.Errorf("%w: invalid entitlement snapshot: %v", ErrBandwidthUnavailable, err)
		}
		var entries []models.UsageLedger
		if err := tx.Where("user_id = ? AND resource_type = ? AND period_key = ?", userID, "bandwidth", record.PeriodKey).Get(&entries); err != nil {
			return fmt.Errorf("%w: usage lookup: %v", ErrBandwidthUnavailable, err)
		}
		var used int64
		for _, entry := range entries {
			used += entry.Delta
		}
		if err := quota.CheckBandwidth(used, bytes, entitlement.Entitlements.MonthlyBandwidthBytes); err != nil {
			if errors.Is(err, quota.ErrQuotaExceeded) {
				return ErrBandwidthQuotaExceeded
			}
			return fmt.Errorf("%w: %v", ErrBandwidthUnavailable, err)
		}
		if _, err := quota.RecordUsageWithQuery(tx, record); err != nil {
			return fmt.Errorf("%w: record: %v", ErrBandwidthUnavailable, err)
		}
		return nil
	})
}
