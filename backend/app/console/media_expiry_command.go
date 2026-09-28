package console

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	mediaservices "goravel/app/services/media"
)

type MediaExpiryCommand struct{}

func (MediaExpiryCommand) Signature() string { return "media:process-expiry" }
func (MediaExpiryCommand) Description() string {
	return "Move expired media to trash and send expiry reminders"
}
func (MediaExpiryCommand) Extend() command.Extend { return command.Extend{Category: "media"} }
func (MediaExpiryCommand) Handle(ctx console.Context) error {
	report, err := mediaservices.ProcessExpiry(context.Background(), time.Now().UTC())
	if err != nil {
		return err
	}
	ctx.Success(fmt.Sprintf("Scanned %d media, expired %d, sent %d reminders, failed %d.", report.Scanned, report.Expired, report.RemindersSent, report.ReminderFailures))
	return nil
}
