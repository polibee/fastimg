package media

import "testing"

func TestValidateAdminMediaOperation(t *testing.T) {
	for _, operation := range []string{"hide", "restore", "approve", "reject", "permanent_delete"} {
		if err := ValidateAdminMediaOperation(operation); err != nil {
			t.Fatalf("operation %q rejected: %v", operation, err)
		}
	}
	if err := ValidateAdminMediaOperation("edit_database_directly"); err == nil {
		t.Fatal("unknown admin media operation should be rejected")
	}
}

func TestModerationStateUpdatesKeepHideSeparateFromReject(t *testing.T) {
	hide := moderationStateUpdates("hide")
	if hide["visibility"] != VisibilityPrivate {
		t.Fatalf("hide visibility = %#v", hide["visibility"])
	}
	if _, exists := hide["moderation_status"]; exists {
		t.Fatal("hide must not mark an image as rejected")
	}
	reject := moderationStateUpdates("reject")
	if reject["visibility"] != VisibilityPrivate || reject["moderation_status"] != ModerationRejected {
		t.Fatalf("reject updates = %#v", reject)
	}
}
