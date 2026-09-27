package providers

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
	billingservices "goravel/app/services/billing"
	mediaservices "goravel/app/services/media"
)

type FastImgQueueProvider struct{}

func (*FastImgQueueProvider) Register(_ contractsfoundation.Application) {}

func (*FastImgQueueProvider) Boot(_ contractsfoundation.Application) {
	facades.Queue().Register([]contractsqueue.Job{
		&billingservices.FulfillOrderJob{},
		&mediaservices.RecoverUploadJob{},
	})
}
