package console

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	planservices "goravel/app/services/plans"
)

type SubscriptionLifecycleCommand struct{}

func (SubscriptionLifecycleCommand) Signature() string { return "subscriptions:process-expiry" }

func (SubscriptionLifecycleCommand) Description() string {
	return "Downgrade expired subscriptions and send expiry reminders"
}

func (SubscriptionLifecycleCommand) Extend() command.Extend {
	return command.Extend{Category: "billing"}
}

func (SubscriptionLifecycleCommand) Handle(ctx console.Context) error {
	report, err := planservices.NewSubscriptionLifecycleService().Process(time.Now().UTC())
	if err != nil {
		return err
	}
	ctx.Success(fmt.Sprintf("Scanned %d subscriptions, expired %d, scheduled %d reminders, sent %d emails, deferred %d, failed %d.", report.Scanned, report.Expired, report.RemindersScheduled, report.EmailsSent, report.EmailsDeferred, report.EmailFailures))
	return nil
}
