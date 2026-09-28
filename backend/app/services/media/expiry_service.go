package media

import (
	"context"
	"fmt"
	"time"

	"goravel/app/facades"
	"goravel/app/models"
	emailservices "goravel/app/services/email"
)

type ExpiryReport struct {
	Scanned          int
	Expired          int
	RemindersSent    int
	ReminderFailures int
}

// ProcessExpiry moves expired media into the user's trash. It never deletes
// objects directly; the existing permanent-delete path remains responsible
// for cleanup and usage release.
func ProcessExpiry(ctx context.Context, now time.Time) (ExpiryReport, error) {
	var report ExpiryReport
	if !facades.Schema().HasColumn("media_assets", "expires_at") {
		return report, nil
	}
	var assets []models.MediaAsset
	if err := facades.Orm().Query().Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", "ready", now).Limit(500).Get(&assets); err != nil {
		return report, err
	}
	report.Scanned = len(assets)
	for _, asset := range assets {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		if _, err := facades.Orm().Query().Model(&models.MediaAsset{}).Where("id = ? AND status = ?", asset.ID, "ready").Update(map[string]any{"status": "deleted", "visibility": "private", "deleted_at": now, "updated_at": now}); err == nil {
			report.Expired++
		}
	}
	var upcoming []models.MediaAsset
	if err := facades.Orm().Query().Where("status = ? AND expires_at > ? AND expires_at <= ? AND expiry_notified_at IS NULL", "ready", now, now.Add(7*24*time.Hour)).Limit(500).Get(&upcoming); err != nil {
		return report, err
	}
	mailer := emailservices.NewService()
	for _, asset := range upcoming {
		var user models.User
		if err := facades.Orm().Query().Where("id = ?", asset.UserID).First(&user); err != nil {
			report.ReminderFailures++
			continue
		}
		err := mailer.Send(ctx, emailservices.Message{To: user.Email, Subject: "Your FastImg image is nearing expiration", Text: fmt.Sprintf("Your image %q will expire on %s.", asset.OriginalName, asset.ExpiresAt.UTC().Format(time.RFC3339)), HTML: fmt.Sprintf("<p>Your image <strong>%s</strong> will expire on %s.</p>", asset.OriginalName, asset.ExpiresAt.UTC().Format(time.RFC3339))})
		if err != nil {
			report.ReminderFailures++
			continue
		}
		if _, err := facades.Orm().Query().Model(&models.MediaAsset{}).Where("id = ? AND expiry_notified_at IS NULL").Update("expiry_notified_at", now); err != nil {
			report.ReminderFailures++
		} else {
			report.RemindersSent++
		}
	}
	return report, nil
}
