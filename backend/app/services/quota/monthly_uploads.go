package quota

import (
	"database/sql"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// CheckMonthlyAPIUploadsWithQuery counts completed uploads and in-flight
// reservations for the current UTC month. Callers must serialize reservations
// for the owner (the upload repository holds that user's row lock).
func CheckMonthlyAPIUploadsWithQuery(query orm.Query, userID uint, now time.Time, limit int64) error {
	if userID == 0 {
		return ErrInvalidUsageRecord
	}
	monthStart := time.Date(now.UTC().Year(), now.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	count, err := query.Model(&models.UploadSession{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND (status = ? OR status = ?)", userID, monthStart, monthEnd, "ready", "processing").
		Count()
	if err != nil {
		return err
	}
	return CheckUploadLimit(count, limit)
}

// CheckMonthlyTransformsWithQuery counts completed transform jobs from the
// ledger and in-flight uploads as one reserved job each. The caller must hold
// the owning user's row lock so concurrent upload reservations are serialized.
func CheckMonthlyTransformsWithQuery(query orm.Query, userID uint, now time.Time, limit int64) error {
	if userID == 0 {
		return ErrInvalidUsageRecord
	}
	if limit == 0 {
		return nil
	}
	monthStart := time.Date(now.UTC().Year(), now.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	periodKey := monthStart.Format("2006-01")
	var completed sql.NullInt64
	if err := query.Model(&models.UsageLedger{}).
		Where("user_id = ? AND resource_type = ? AND period_key = ?", userID, "transform", periodKey).
		Sum("delta", &completed); err != nil {
		return err
	}
	completedValue := nullableSumInt64(completed)
	if completedValue < 0 {
		completedValue = 0
	}
	pending, err := query.Model(&models.UploadSession{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND status = ?", userID, monthStart, monthEnd, "processing").
		Count()
	if err != nil {
		return err
	}
	if err := CheckUsageLimit(completedValue, pending, limit); err != nil {
		return err
	}
	return CheckUploadLimit(completedValue+pending, limit)
}

func nullableSumInt64(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}
