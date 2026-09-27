package console

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/facades"
	mediaservices "goravel/app/services/media"
)

type MediaRecoveryDispatchCommand struct{}

func (MediaRecoveryDispatchCommand) Signature() string { return "media:dispatch-recovery" }
func (MediaRecoveryDispatchCommand) Description() string {
	return "Dispatch stale image processing sessions to the Redis queue"
}

func (MediaRecoveryDispatchCommand) Extend() command.Extend { return command.Extend{Category: "media"} }

func (MediaRecoveryDispatchCommand) Handle(ctx console.Context) error {
	if !facades.Schema().HasTable("upload_sessions") {
		ctx.Success("No upload session table is available.")
		return nil
	}
	var rows []struct {
		ID     uint `json:"id"`
		UserID uint `json:"user_id"`
	}
	cutoff := time.Now().UTC().Add(-10 * time.Minute)
	if err := facades.Orm().Query().Table("upload_sessions").Select("id", "user_id").Where("status = ? AND updated_at < ?", "processing", cutoff).OrderBy("id", "asc").Limit(100).Get(&rows); err != nil {
		return err
	}
	queued := 0
	for _, row := range rows {
		if err := mediaservices.DispatchUploadRecovery(row.UserID, row.ID); err != nil {
			return fmt.Errorf("dispatch upload session %d: %w", row.ID, err)
		}
		queued++
	}
	ctx.Success(fmt.Sprintf("Dispatched %d stale upload recovery jobs.", queued))
	return nil
}
