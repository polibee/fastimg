package models

import (
	"encoding/json"
	"testing"
)

func TestUsageLedgerJSONDoesNotExposeIdempotencyKey(t *testing.T) {
	encoded, err := json.Marshal(UsageLedger{IdempotencyKey: "internal-key"})
	if err != nil {
		t.Fatalf("marshal UsageLedger: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("decode UsageLedger: %v", err)
	}
	if _, exists := fields["idempotency_key"]; exists {
		t.Fatalf("internal idempotency key leaked in API serialization: %s", encoded)
	}
}
