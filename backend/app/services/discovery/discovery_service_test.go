package discovery

import "testing"

func TestNormalizePageBoundsPublicFeedRequests(t *testing.T) {
	tests := []struct {
		name                  string
		page, perPage         int
		wantPage, wantPerPage int
	}{
		{name: "defaults", page: 0, perPage: 0, wantPage: 1, wantPerPage: 24},
		{name: "caps oversized page", page: 3, perPage: 999, wantPage: 3, wantPerPage: 48},
		{name: "preserves valid values", page: 2, perPage: 12, wantPage: 2, wantPerPage: 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, perPage := NormalizePage(tt.page, tt.perPage)
			if page != tt.wantPage || perPage != tt.wantPerPage {
				t.Fatalf("NormalizePage(%d, %d) = (%d, %d), want (%d, %d)", tt.page, tt.perPage, page, perPage, tt.wantPage, tt.wantPerPage)
			}
		})
	}
}

func TestPublicDiscoveryOnlyAcceptsKnownVariants(t *testing.T) {
	for _, variant := range []string{"original", "thumbnail", "medium"} {
		if !validVariant(variant) {
			t.Fatalf("variant %q should be public", variant)
		}
	}
	for _, variant := range []string{"", "avatar", "../../private"} {
		if validVariant(variant) {
			t.Fatalf("variant %q should be rejected", variant)
		}
	}
}
