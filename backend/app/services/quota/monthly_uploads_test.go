package quota

import (
	"database/sql"
	"testing"
)

func TestNullableSumInt64TreatsEmptyAggregateAsZero(t *testing.T) {
	if got := nullableSumInt64(sql.NullInt64{}); got != 0 {
		t.Fatalf("nullableSumInt64(empty) = %d, want 0", got)
	}
	if got := nullableSumInt64(sql.NullInt64{Int64: 7, Valid: true}); got != 7 {
		t.Fatalf("nullableSumInt64(value) = %d, want 7", got)
	}
}
