package quota

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrStorageQuotaExceeded = errors.New("storage quota exceeded")
	ErrFileTooLarge         = errors.New("file exceeds plan limit")
	ErrInvalidUploadSize    = errors.New("upload size must be positive")
	ErrInvalidEntitlement   = errors.New("invalid entitlement")
	ErrQuotaExceeded        = errors.New("usage quota exceeded")
	ErrInvalidUsageAmount   = errors.New("usage amount cannot be negative")
)

// Entitlement contains server-side limits for one effective subscription snapshot.
// Values are defaults for seeding only; production reads the persisted snapshot.
type Entitlement struct {
	StorageBytes          int64 `json:"storage_bytes"`
	MaxFileBytes          int64 `json:"max_file_bytes"`
	DailyUploads          int64 `json:"daily_uploads"`
	MonthlyAPIUploads     int64 `json:"monthly_api_uploads"`
	MonthlyBandwidthBytes int64 `json:"monthly_bandwidth_bytes"`
	TransformCount        int64 `json:"transform_count"`
	APIRatePerMinute      int64 `json:"api_rate_per_minute"`
	TokenLimit            int64 `json:"token_limit"`
	AdsEnabled            bool  `json:"ads_enabled"`
	WatermarkEnabled      bool  `json:"watermark_enabled"`
}

// ParseEntitlementJSON validates a complete entitlement object. PascalCase keys
// from snapshots written before JSON tags were added remain readable.
func ParseEntitlementJSON(payload string) (Entitlement, error) {
	decoder := json.NewDecoder(strings.NewReader(payload))
	var fields map[string]json.RawMessage
	if err := decoder.Decode(&fields); err != nil || fields == nil {
		return Entitlement{}, fmt.Errorf("%w: expected a JSON object", ErrInvalidEntitlement)
	}
	if _, versioned := fields["schema_version"]; versioned {
		var snapshot struct {
			SchemaVersion int             `json:"schema_version"`
			PlanVersion   string          `json:"plan_version"`
			Plan          json.RawMessage `json:"plan"`
			Entitlements  json.RawMessage `json:"entitlements"`
		}
		if err := json.Unmarshal([]byte(payload), &snapshot); err != nil || snapshot.SchemaVersion != 1 || snapshot.PlanVersion == "" || len(snapshot.Plan) == 0 || len(snapshot.Entitlements) == 0 {
			return Entitlement{}, fmt.Errorf("%w: invalid versioned subscription snapshot", ErrInvalidEntitlement)
		}
		for key := range fields {
			if key != "schema_version" && key != "plan_version" && key != "plan" && key != "entitlements" {
				return Entitlement{}, fmt.Errorf("%w: unknown snapshot field %q", ErrInvalidEntitlement, key)
			}
		}
		return ParseEntitlementJSON(string(snapshot.Entitlements))
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Entitlement{}, fmt.Errorf("%w: trailing JSON value", ErrInvalidEntitlement)
	}

	canonical := make(map[string]json.RawMessage, len(fields))
	for name, value := range fields {
		canonicalName, ok := canonicalEntitlementField(name)
		if !ok {
			return Entitlement{}, fmt.Errorf("%w: unknown field %q", ErrInvalidEntitlement, name)
		}
		if _, exists := canonical[canonicalName]; exists {
			return Entitlement{}, fmt.Errorf("%w: duplicate field %q", ErrInvalidEntitlement, canonicalName)
		}
		canonical[canonicalName] = value
	}

	for _, name := range []string{
		"storage_bytes", "max_file_bytes", "daily_uploads", "monthly_api_uploads",
		"monthly_bandwidth_bytes", "transform_count", "api_rate_per_minute",
		"token_limit", "ads_enabled",
	} {
		if _, ok := canonical[name]; !ok {
			return Entitlement{}, fmt.Errorf("%w: missing field %q", ErrInvalidEntitlement, name)
		}
	}

	encoded, err := json.Marshal(canonical)
	if err != nil {
		return Entitlement{}, fmt.Errorf("%w: encode fields: %v", ErrInvalidEntitlement, err)
	}
	var entitlement Entitlement
	if err := json.Unmarshal(encoded, &entitlement); err != nil {
		return Entitlement{}, fmt.Errorf("%w: decode fields: %v", ErrInvalidEntitlement, err)
	}
	if err := entitlement.Validate(); err != nil {
		return Entitlement{}, err
	}
	return entitlement, nil
}

func canonicalEntitlementField(name string) (string, bool) {
	switch name {
	case "storage_bytes", "StorageBytes":
		return "storage_bytes", true
	case "max_file_bytes", "MaxFileBytes":
		return "max_file_bytes", true
	case "daily_uploads", "DailyUploads":
		return "daily_uploads", true
	case "monthly_api_uploads", "MonthlyAPIUploads":
		return "monthly_api_uploads", true
	case "monthly_bandwidth_bytes", "MonthlyBandwidthBytes":
		return "monthly_bandwidth_bytes", true
	case "transform_count", "TransformCount":
		return "transform_count", true
	case "api_rate_per_minute", "APIRatePerMinute":
		return "api_rate_per_minute", true
	case "token_limit", "TokenLimit":
		return "token_limit", true
	case "ads_enabled", "AdsEnabled":
		return "ads_enabled", true
	case "watermark_enabled", "WatermarkEnabled":
		return "watermark_enabled", true
	default:
		return "", false
	}
}

func (e Entitlement) Validate() error {
	if e.StorageBytes <= 0 {
		return fmt.Errorf("%w: storage_bytes must be positive", ErrInvalidEntitlement)
	}
	for name, value := range map[string]int64{
		"max_file_bytes":          e.MaxFileBytes,
		"daily_uploads":           e.DailyUploads,
		"monthly_api_uploads":     e.MonthlyAPIUploads,
		"monthly_bandwidth_bytes": e.MonthlyBandwidthBytes,
		"transform_count":         e.TransformCount,
		"api_rate_per_minute":     e.APIRatePerMinute,
		"token_limit":             e.TokenLimit,
	} {
		if value < 0 {
			return fmt.Errorf("%w: %s cannot be negative", ErrInvalidEntitlement, name)
		}
	}
	return nil
}

type UsageState struct {
	StorageBytes  int64
	ReservedBytes int64
}

type Reservation struct {
	Bytes          int64
	RemainingBytes int64
}

type UsageEntry struct {
	ResourceType string
	Delta        int64
	PeriodKey    string
}

// AggregateUsage treats storage as a lifetime balance and usage counters as
// period-scoped. storage_bytes is accepted as a legacy name for storage.
func AggregateUsage(entries []UsageEntry, periodKey string) map[string]int64 {
	totals := map[string]int64{
		"storage": 0, "bandwidth": 0, "upload": 0, "api": 0, "transform": 0,
	}
	for _, entry := range entries {
		resource := entry.ResourceType
		if resource == "storage_bytes" {
			resource = "storage"
		}
		if _, supported := totals[resource]; !supported {
			continue
		}
		if resource != "storage" && entry.PeriodKey != periodKey {
			continue
		}
		totals[resource] += entry.Delta
	}
	return totals
}

func CheckUploadLimit(used, limit int64) error {
	return CheckUsageLimit(used, 1, limit)
}

func CheckBandwidth(usedBytes, requestedBytes, limitBytes int64) error {
	return CheckUsageLimit(usedBytes, requestedBytes, limitBytes)
}

// CheckUsageLimit treats a zero limit as unlimited and avoids addition
// overflow by comparing the request to the remaining allowance.
func CheckUsageLimit(used, requested, limit int64) error {
	if used < 0 || requested < 0 || limit < 0 {
		return ErrInvalidUsageAmount
	}
	if limit == 0 {
		return nil
	}
	if used > limit || requested > limit-used {
		return ErrQuotaExceeded
	}
	return nil
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
		WatermarkEnabled:      false,
	}
}

func ReserveStorage(state UsageState, entitlement Entitlement, bytes int64) (Reservation, error) {
	if bytes <= 0 {
		return Reservation{}, ErrInvalidUploadSize
	}
	if state.StorageBytes < 0 || state.ReservedBytes < 0 {
		return Reservation{}, ErrStorageQuotaExceeded
	}
	if entitlement.MaxFileBytes > 0 && bytes > entitlement.MaxFileBytes {
		return Reservation{}, ErrFileTooLarge
	}
	if entitlement.StorageBytes <= 0 || state.StorageBytes > entitlement.StorageBytes || state.ReservedBytes > entitlement.StorageBytes-state.StorageBytes {
		return Reservation{}, ErrStorageQuotaExceeded
	}
	used := state.StorageBytes + state.ReservedBytes
	if bytes > entitlement.StorageBytes-used {
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
