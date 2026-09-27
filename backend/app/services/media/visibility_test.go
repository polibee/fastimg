package media

import "testing"

func TestDefaultMediaVisibilityIsPublic(t *testing.T) {
	if got := DefaultVisibility; got != VisibilityPublic {
		t.Fatalf("default visibility = %q, want %q", got, VisibilityPublic)
	}
	if got := DefaultModerationStatus; got != ModerationApproved {
		t.Fatalf("default moderation status = %q, want %q", got, ModerationApproved)
	}
}

func TestPublicDeliveryRequiresVisibleAndNonRejectedMedia(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		moderation string
		want       bool
	}{
		{name: "public pending upload is available", visibility: VisibilityPublic, moderation: ModerationPending, want: true},
		{name: "unlisted is available by explicit link", visibility: VisibilityLink, moderation: ModerationApproved, want: true},
		{name: "private is owner only", visibility: VisibilityPrivate, moderation: ModerationApproved, want: false},
		{name: "rejected is never public", visibility: VisibilityPublic, moderation: ModerationRejected, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AllowsPublicDelivery(tt.visibility, tt.moderation); got != tt.want {
				t.Fatalf("AllowsPublicDelivery(%q, %q) = %v, want %v", tt.visibility, tt.moderation, got, tt.want)
			}
		})
	}
}

func TestSignedDeliveryNeverBypassesRejectedMedia(t *testing.T) {
	if AllowsSignedDelivery(VisibilityPrivate, ModerationRejected, false, true) {
		t.Fatal("rejected private media must not be delivered by a temporary signed URL")
	}
	if !AllowsSignedDelivery(VisibilityPrivate, ModerationApproved, false, true) {
		t.Fatal("an owner-created temporary signed URL may deliver approved private media")
	}
}
