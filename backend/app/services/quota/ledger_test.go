package quota

import (
	"errors"
	"testing"
)

type memoryUsageLedgerStore struct {
	records map[string]UsageRecord
}

func (s *memoryUsageLedgerStore) FindUsageByIdempotencyKey(key string) (UsageRecord, bool, error) {
	record, exists := s.records[key]
	return record, exists, nil
}

func (s *memoryUsageLedgerStore) CreateUsageRecord(record UsageRecord) error {
	if s.records == nil {
		s.records = make(map[string]UsageRecord)
	}
	s.records[record.IdempotencyKey] = record
	return nil
}

func TestRecordUsageOnceIsIdempotentAndRejectsChangedPayload(t *testing.T) {
	store := &memoryUsageLedgerStore{}
	input := UsageRecord{UserID: 17, ResourceType: "storage_bytes", Delta: 512, SourceType: "upload", SourceID: "session:42", PeriodKey: "lifetime"}

	created, err := RecordUsageOnce(store, input)
	if err != nil || !created {
		t.Fatalf("first RecordUsageOnce() = (%v, %v), want (true, nil)", created, err)
	}
	created, err = RecordUsageOnce(store, input)
	if err != nil || created || len(store.records) != 1 {
		t.Fatalf("duplicate RecordUsageOnce() = (%v, %v), records=%d; want (false, nil), one record", created, err, len(store.records))
	}
	for _, stored := range store.records {
		if stored.ResourceType != "storage" {
			t.Fatalf("legacy storage resource name was not canonicalized: %+v", stored)
		}
	}

	input.Delta = 513
	if _, err := RecordUsageOnce(store, input); !errors.Is(err, ErrUsageIdempotencyConflict) {
		t.Fatalf("changed duplicate error = %v, want ErrUsageIdempotencyConflict", err)
	}
}

func TestRecordUsageOnceSeparatesUsersAndValidatesPeriod(t *testing.T) {
	store := &memoryUsageLedgerStore{}
	input := UsageRecord{UserID: 17, ResourceType: "api", Delta: 1, SourceType: "api_request", SourceID: "request:42", PeriodKey: "2026-09"}
	if created, err := RecordUsageOnce(store, input); err != nil || !created {
		t.Fatalf("first API usage = (%v, %v), want (true, nil)", created, err)
	}
	input.UserID = 18
	if created, err := RecordUsageOnce(store, input); err != nil || !created || len(store.records) != 2 {
		t.Fatalf("other user's usage = (%v, %v), records=%d; want separate record", created, err, len(store.records))
	}
	input.PeriodKey = "2026-13"
	if _, err := RecordUsageOnce(store, input); !errors.Is(err, ErrInvalidUsageRecord) {
		t.Fatalf("invalid period error = %v, want ErrInvalidUsageRecord", err)
	}
}

func TestRecordUsageOnceRejectsWrongSignForUsageSource(t *testing.T) {
	store := &memoryUsageLedgerStore{}
	cases := []UsageRecord{
		{UserID: 1, ResourceType: "storage", Delta: 100, SourceType: "delete", SourceID: "media:2", PeriodKey: "lifetime"},
		{UserID: 1, ResourceType: "storage", Delta: -100, SourceType: "restore", SourceID: "media:2", PeriodKey: "lifetime"},
		{UserID: 1, ResourceType: "api", Delta: -1, SourceType: "api_request", SourceID: "request:2", PeriodKey: "2026-09"},
	}
	for _, input := range cases {
		if _, err := RecordUsageOnce(store, input); !errors.Is(err, ErrInvalidUsageRecord) {
			t.Errorf("RecordUsageOnce(%+v) error = %v, want ErrInvalidUsageRecord", input, err)
		}
	}
}
