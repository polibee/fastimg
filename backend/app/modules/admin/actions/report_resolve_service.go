package actions

import (
	"github.com/goravel/framework/contracts/database/orm"
	"goravel/app/facades"
	audit "goravel/app/services/audit"
	"time"
)

type reportRow struct {
	ID      uint   `db:"id"`
	MediaID uint   `db:"media_asset_id"`
	Status  string `db:"status"`
}

func resolveReport(id, op uint, act, n string) error {
	now := time.Now()
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var r reportRow
		if e := tx.Table("media_reports").Where("id = ?", id).LockForUpdate().First(&r); e != nil || r.ID == 0 {
			return ErrReportNotFound
		}
		if r.Status != "pending" && r.Status != "in_review" {
			return ErrReportAlreadyResolved
		}
		if act == "hide_media" {
			if _, e := tx.Table("media_assets").Where("id = ?", r.MediaID).Update(map[string]any{"visibility": "private", "moderation_status": "rejected", "updated_at": now}); e != nil {
				return e
			}
		}
		if act == "restore_media" {
			if _, e := tx.Table("media_assets").Where("id = ?", r.MediaID).Update(map[string]any{"visibility": "public", "moderation_status": "approved", "updated_at": now}); e != nil {
				return e
			}
		}
		if _, e := tx.Table("media_reports").Where("id = ?", id).Update(map[string]any{"status": "resolved", "resolution": n, "resolved_by": op, "resolved_at": now, "updated_at": now}); e != nil {
			return e
		}
		_ = audit.NewAuditService().Record(op, "moderation.report.resolve", map[string]any{"report_id": id, "media_asset_id": r.MediaID, "action": act})
		return nil
	})
}
