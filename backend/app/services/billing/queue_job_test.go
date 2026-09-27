package billing

import (
	"testing"
	"time"
)

func TestFulfillOrderJobContract(t *testing.T) {
	job := &FulfillOrderJob{}
	if job.Signature() != FulfillOrderJobSignature {
		t.Fatalf("signature = %q", job.Signature())
	}
	if err := job.Handle(); err == nil {
		t.Fatal("expected missing order id error")
	}
	if retry, delay := job.ShouldRetry(nil, 3); retry || delay != 0 {
		t.Fatalf("final retry = %v, %s", retry, delay)
	}
	if retry, delay := job.ShouldRetry(nil, 1); !retry || delay != 60*time.Second {
		t.Fatalf("second attempt retry = %v, %s", retry, delay)
	}
}

func TestDispatchFulfillmentRejectsEmptyOrder(t *testing.T) {
	if err := DispatchFulfillment(0); err == nil {
		t.Fatal("expected empty order id error")
	}
}
