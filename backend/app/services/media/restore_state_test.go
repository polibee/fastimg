package media

import "testing"

func TestCanRestoreOnlyMediaThatHasNotEnteredPhysicalCleanup(t *testing.T) {
	for _, status := range []string{"ready", "processing", "cleanup_pending", "physically_deleted", "failed"} {
		if canRestoreMediaStatus(status) {
			t.Errorf("canRestoreMediaStatus(%q) = true, want false", status)
		}
	}
	if !canRestoreMediaStatus("deleted") {
		t.Fatal(`canRestoreMediaStatus("deleted") = false, want true`)
	}
}

func TestCleanupPendingMediaRemainsVisibleInTrashForRetry(t *testing.T) {
	statuses := mediaTrashStatuses()
	if len(statuses) != 2 || statuses[0] != "deleted" || statuses[1] != "cleanup_pending" {
		t.Fatalf("mediaTrashStatuses() = %v, want deleted and cleanup_pending", statuses)
	}
}
