package billing

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
)

const FulfillOrderJobSignature = "fastimg.billing.fulfill-order"

type FulfillOrderJob struct{}

func (*FulfillOrderJob) Signature() string { return FulfillOrderJobSignature }

func (*FulfillOrderJob) Handle(args ...any) error {
	if len(args) == 0 {
		return fmt.Errorf("order id is required")
	}
	orderID, ok := args[0].(uint)
	if !ok || orderID == 0 {
		return fmt.Errorf("invalid order id")
	}
	return NewFulfillmentService().Process(orderID)
}

func (*FulfillOrderJob) ShouldRetry(_ error, attempt int) (bool, time.Duration) {
	if attempt >= 3 {
		return false, 0
	}
	return true, time.Duration(attempt+1) * 30 * time.Second
}

func DispatchFulfillment(orderID uint) error {
	if orderID == 0 {
		return ErrFulfillmentNotReady
	}
	return facades.Queue().Job(&FulfillOrderJob{}, []queue.Arg{{Value: orderID, Type: "uint"}}).OnQueue("billing").Dispatch()
}
