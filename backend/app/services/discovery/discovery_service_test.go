package discovery

import (
	"context"
	"testing"

	storageservices "goravel/app/services/storage"
)

func TestServiceResolvesStoragePerContentRequest(t *testing.T) {
	want := &storageservices.LocalProvider{}
	service := NewServiceWithStorageResolver(nil, func(context.Context) (storageservices.StorageProvider, error) {
		return want, nil
	})

	got, err := service.resolveStorage(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("resolved storage = %p, want %p", got, want)
	}
}

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

func TestIsDiscoverableUsesPostModerationRule(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		moderation string
		want       bool
	}{
		{name: "new approved upload", visibility: VisibilityPublic, moderation: ModerationApproved, want: true},
		{name: "public upload awaiting post moderation", visibility: VisibilityPublic, moderation: ModerationPending, want: true},
		{name: "public upload under review", visibility: VisibilityPublic, moderation: "manual_review", want: true},
		{name: "link only media", visibility: VisibilityLink, moderation: ModerationApproved, want: false},
		{name: "private media", visibility: VisibilityPrivate, moderation: ModerationApproved, want: false},
		{name: "rejected public media", visibility: VisibilityPublic, moderation: ModerationRejected, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDiscoverable(tt.visibility, tt.moderation); got != tt.want {
				t.Fatalf("IsDiscoverable(%q, %q) = %v, want %v", tt.visibility, tt.moderation, got, tt.want)
			}
		})
	}
}
