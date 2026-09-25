package migrations

import (
	"reflect"
	"testing"
)

func TestMissingRBACTimestampColumnsSkipsExistingColumns(t *testing.T) {
	existing := map[string]map[string]bool{
		"roles":       {"created_at": true, "updated_at": true},
		"permissions": {"created_at": true},
	}
	hasColumn := func(table, column string) bool {
		return existing[table][column]
	}

	if got := missingRBACTimestampColumns("roles", hasColumn); len(got) != 0 {
		t.Fatalf("expected no missing role columns, got %v", got)
	}
	if got, want := missingRBACTimestampColumns("permissions", hasColumn), []string{"updated_at"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("missing permissions columns = %v, want %v", got, want)
	}
}
