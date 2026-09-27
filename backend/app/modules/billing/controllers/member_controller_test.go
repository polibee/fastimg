package controllers

import (
	"testing"

	"goravel/app/services/billing"
)

func TestBillingErrorCodeMapsUncancellableOrder(t *testing.T) {
	code, status := billingErrorCode(billing.ErrOrderAlreadyCanceled)
	if code != "ORDER_NOT_CANCELLABLE" || status != 409 {
		t.Fatalf("billingErrorCode() = %q, %d; want ORDER_NOT_CANCELLABLE, 409", code, status)
	}
}
