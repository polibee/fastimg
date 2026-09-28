package migrations

import "testing"

func TestContentNavigationFriendLinksMigrationHasStableSignature(t *testing.T) {
	migration := &M20260928000001CreateContentNavigationFriendLinksTables{}
	if got, want := migration.Signature(), "20260928000001_create_content_navigation_friend_links_tables"; got != want {
		t.Fatalf("migration signature = %q, want %q", got, want)
	}
}
