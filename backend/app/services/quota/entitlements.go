package quota

import "errors"

var (
	ErrStorageQuotaExceeded = errors.New("storage quota exceeded")
	ErrFileTooLarge         = errors.New("file exceeds plan limit")
	ErrInvalidUploadSize    = errors.New("upload size must be positive")
)

// Entitlement contains server-side limits for one effective subscription snapshot.
// Values are defaults for seeding only; production reads the persisted snapshot.
type Entitlement struct {
	StorageBytes          int64
	MaxFileBytes          int64
	DailyUploads          int64
	MonthlyAPIUploads     int64
	MonthlyBandwidthBytes int64
	TransformCount        int64
	APIRatePerMinute      int64
	TokenLimit            int64
	AdsEnabled            bool
}

type UsageState struct {
	StorageBytes  int64
	ReservedBytes int64
}

type Reservation struct {
	Bytes          int64
	RemainingBytes int64
}

func DefaultFreeEntitlement() Entitlement {
	return Entitlement{
		StorageBytes:          1_000_000_000,
		MaxFileBytes:          10_000_000,
		DailyUploads:          100,
		MonthlyAPIUploads:     500,
		MonthlyBandwidthBytes: 5_000_000_000,
		TransformCount:        500,
		APIRatePerMinute:      30,
		TokenLimit:            1,
		AdsEnabled:            true,
	}
}

func ReserveStorage(state UsageState, entitlement Entitlement, bytes int64) (Reservation, error) {
	if bytes <= 0 {
		return Reservation{}, ErrInvalidUploadSize
	}
	if entitlement.MaxFileBytes > 0 && bytes > entitlement.MaxFileBytes {
		return Reservation{}, ErrFileTooLarge
	}
	used := state.StorageBytes + state.ReservedBytes
	if entitlement.StorageBytes <= 0 || used > entitlement.StorageBytes || bytes > entitlement.StorageBytes-used {
		return Reservation{}, ErrStorageQuotaExceeded
	}
	return Reservation{
		Bytes:          bytes,
		RemainingBytes: entitlement.StorageBytes - used - bytes,
	}, nil
}

func ReleaseStorage(state UsageState, bytes int64) UsageState {
	if bytes <= 0 {
		return state
	}
	if bytes >= state.StorageBytes {
		state.StorageBytes = 0
	} else {
		state.StorageBytes -= bytes
	}
	if bytes >= state.ReservedBytes {
		state.ReservedBytes = 0
	} else {
		state.ReservedBytes -= bytes
	}
	return state
}
