package quota

import (
	"testing"
)

func TestFreeEntitlementProvidesUsableBaseline(t *testing.T) {
	got := DefaultFreeEntitlement()

	if got.StorageBytes != 1_000_000_000 {
		t.Fatalf("storage = %d, want 1000000000", got.StorageBytes)
	}
	if got.MaxFileBytes != 10_000_000 {
		t.Fatalf("max file = %d, want 10000000", got.MaxFileBytes)
	}
	if got.DailyUploads != 100 || got.MonthlyAPIUploads != 500 {
		t.Fatalf("upload limits = %+v, want daily=100 monthly=500", got)
	}
	if !got.AdsEnabled {
		t.Fatal("free entitlement should allow ads")
	}
}

func TestReserveStorageRejectsLimitAndTracksReservation(t *testing.T) {
	state := UsageState{StorageBytes: 900, ReservedBytes: 50}
	limit := Entitlement{StorageBytes: 1_000, MaxFileBytes: 600}

	reservation, err := ReserveStorage(state, limit, 50)
	if err != nil {
		t.Fatalf("reserve storage: %v", err)
	}
	if reservation.Bytes != 50 || reservation.RemainingBytes != 0 {
		t.Fatalf("reservation = %+v", reservation)
	}

	if _, err := ReserveStorage(state, limit, 51); err != ErrStorageQuotaExceeded {
		t.Fatalf("expected storage quota error, got %v", err)
	}
}

func TestReserveStorageRejectsSingleFileLimit(t *testing.T) {
	_, err := ReserveStorage(UsageState{}, Entitlement{StorageBytes: 1_000, MaxFileBytes: 100}, 101)
	if err != ErrFileTooLarge {
		t.Fatalf("expected file too large, got %v", err)
	}
}

func TestReleaseStorageNeverCreatesNegativeReservation(t *testing.T) {
	state := UsageState{StorageBytes: 100, ReservedBytes: 20}
	got := ReleaseStorage(state, 50)
	if got.StorageBytes != 50 || got.ReservedBytes != 0 {
		t.Fatalf("released state = %+v", got)
	}
}
