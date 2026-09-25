package quota

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

var (
	ErrInvalidUsageRecord       = errors.New("invalid usage record")
	ErrUsageIdempotencyConflict = errors.New("usage idempotency key conflicts with an existing record")
)

type UsageRecord struct {
	UserID         uint
	ResourceType   string
	Delta          int64
	SourceType     string
	SourceID       string
	PeriodKey      string
	IdempotencyKey string
}

type UsageLedgerStore interface {
	FindUsageByIdempotencyKey(key string) (UsageRecord, bool, error)
	CreateUsageRecord(record UsageRecord) error
}

// RecordUsageOnce validates and writes a usage event once per stable source.
// Callers using database transactions must serialize writes for a user (the
// upload and media lifecycle paths lock that user's row) in addition to the
// database unique idempotency index.
func RecordUsageOnce(store UsageLedgerStore, input UsageRecord) (bool, error) {
	if store == nil {
		return false, fmt.Errorf("%w: usage store is required", ErrInvalidUsageRecord)
	}
	record, err := normalizeUsageRecord(input)
	if err != nil {
		return false, err
	}
	record.IdempotencyKey, err = usageIdempotencyKey(record)
	if err != nil {
		return false, err
	}
	existing, found, err := store.FindUsageByIdempotencyKey(record.IdempotencyKey)
	if err != nil {
		return false, err
	}
	if found {
		if sameUsageRecord(existing, record) {
			return false, nil
		}
		return false, ErrUsageIdempotencyConflict
	}
	if err := store.CreateUsageRecord(record); err != nil {
		return false, err
	}
	return true, nil
}

// RecordUsageWithQuery writes through the caller's transaction. The transaction
// must hold the owning user's row lock to serialize concurrent source checks.
func RecordUsageWithQuery(query orm.Query, input UsageRecord) (bool, error) {
	return RecordUsageOnce(queryUsageLedgerStore{query: query}, input)
}

func normalizeUsageRecord(record UsageRecord) (UsageRecord, error) {
	if record.UserID == 0 || record.Delta == 0 || strings.TrimSpace(record.SourceID) != record.SourceID || record.SourceID == "" || len(record.SourceID) > 120 {
		return UsageRecord{}, fmt.Errorf("%w: missing or invalid identity/delta", ErrInvalidUsageRecord)
	}
	if record.ResourceType == "storage_bytes" {
		record.ResourceType = "storage"
	}
	switch record.ResourceType {
	case "storage", "bandwidth", "upload", "api", "transform":
	default:
		return UsageRecord{}, fmt.Errorf("%w: unsupported resource type", ErrInvalidUsageRecord)
	}
	switch record.SourceType {
	case "upload", "delete", "restore", "download", "api_request", "adjustment":
	default:
		return UsageRecord{}, fmt.Errorf("%w: unsupported source type", ErrInvalidUsageRecord)
	}
	switch record.SourceType {
	case "upload", "restore", "download", "api_request":
		if record.Delta < 0 {
			return UsageRecord{}, fmt.Errorf("%w: source requires a positive delta", ErrInvalidUsageRecord)
		}
	case "delete":
		if record.Delta > 0 {
			return UsageRecord{}, fmt.Errorf("%w: delete requires a negative delta", ErrInvalidUsageRecord)
		}
	}
	if record.PeriodKey == "" || len(record.PeriodKey) > 32 {
		return UsageRecord{}, fmt.Errorf("%w: period key is required", ErrInvalidUsageRecord)
	}
	if record.ResourceType == "storage" {
		if record.PeriodKey != "lifetime" {
			return UsageRecord{}, fmt.Errorf("%w: storage period must be lifetime", ErrInvalidUsageRecord)
		}
	} else {
		period, err := time.Parse("2006-01", record.PeriodKey)
		if err != nil || period.Format("2006-01") != record.PeriodKey {
			return UsageRecord{}, fmt.Errorf("%w: period must be YYYY-MM", ErrInvalidUsageRecord)
		}
	}
	return record, nil
}

func usageIdempotencyKey(record UsageRecord) (string, error) {
	identity, err := json.Marshal(struct {
		Version      int    `json:"v"`
		UserID       uint   `json:"user_id"`
		ResourceType string `json:"resource_type"`
		SourceType   string `json:"source_type"`
		SourceID     string `json:"source_id"`
		PeriodKey    string `json:"period_key"`
	}{1, record.UserID, record.ResourceType, record.SourceType, record.SourceID, record.PeriodKey})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(identity)
	return "usage:v1:" + hex.EncodeToString(digest[:]), nil
}

func sameUsageRecord(left, right UsageRecord) bool {
	return left.UserID == right.UserID && left.ResourceType == right.ResourceType &&
		left.Delta == right.Delta && left.SourceType == right.SourceType &&
		left.SourceID == right.SourceID && left.PeriodKey == right.PeriodKey
}

type queryUsageLedgerStore struct {
	query orm.Query
}

func (s queryUsageLedgerStore) FindUsageByIdempotencyKey(key string) (UsageRecord, bool, error) {
	query := s.query.Model(&models.UsageLedger{}).Where("idempotency_key = ?", key)
	exists, err := query.Exists()
	if err != nil || !exists {
		return UsageRecord{}, exists, err
	}
	var stored models.UsageLedger
	if err := query.First(&stored); err != nil {
		return UsageRecord{}, false, err
	}
	return UsageRecord{
		UserID: stored.UserID, ResourceType: stored.ResourceType, Delta: stored.Delta,
		SourceType: stored.SourceType, SourceID: stored.SourceID,
		PeriodKey: stored.PeriodKey, IdempotencyKey: stored.IdempotencyKey,
	}, true, nil
}

func (s queryUsageLedgerStore) CreateUsageRecord(record UsageRecord) error {
	return s.query.Create(&models.UsageLedger{
		UserID: record.UserID, ResourceType: record.ResourceType, Delta: record.Delta,
		SourceType: record.SourceType, SourceID: record.SourceID,
		IdempotencyKey: record.IdempotencyKey, PeriodKey: record.PeriodKey,
	})
}
