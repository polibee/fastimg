package controllers

import (
	"errors"
	"testing"

	"goravel/app/services/quota"
)

func TestResolveWatermarkEntitlementDoesNotBlockAdministrators(t *testing.T) {
	watermarkEnabled, err := ResolveWatermarkEntitlement(true, quota.Entitlement{}, errors.New("subscription unavailable"))
	if err != nil || watermarkEnabled {
		t.Fatalf("administrator fallback = %v, %v; want disabled watermark without an upload block", watermarkEnabled, err)
	}

	watermarkEnabled, err = ResolveWatermarkEntitlement(false, quota.Entitlement{}, errors.New("subscription unavailable"))
	if err == nil || watermarkEnabled {
		t.Fatal("member subscription errors must remain visible")
	}
}
