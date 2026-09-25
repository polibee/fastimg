package quota

import "testing"

func TestNormalizeUsageLedgerPage(t *testing.T) {
	tests := []struct {
		name                  string
		page, perPage         int
		wantPage, wantPerPage int
	}{
		{name: "defaults", page: 0, perPage: 0, wantPage: 1, wantPerPage: 20},
		{name: "negative values", page: -3, perPage: -8, wantPage: 1, wantPerPage: 20},
		{name: "maximum page size", page: 4, perPage: 1000, wantPage: 4, wantPerPage: 100},
		{name: "preserve valid values", page: 2, perPage: 50, wantPage: 2, wantPerPage: 50},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			page, perPage := normalizeUsageLedgerPage(test.page, test.perPage)
			if page != test.wantPage || perPage != test.wantPerPage {
				t.Fatalf("normalizeUsageLedgerPage(%d, %d) = (%d, %d), want (%d, %d)", test.page, test.perPage, page, perPage, test.wantPage, test.wantPerPage)
			}
		})
	}
}
