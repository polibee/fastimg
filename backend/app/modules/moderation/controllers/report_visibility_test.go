package controllers

import "testing"

func TestReportablePublicMediaMatchesPostModerationDiscovery(t *testing.T) {
	for _, status := range []string{"pending", "manual_review", "approved"} {
		if !ReportablePublicMedia(status) {
			t.Fatalf("%s public media should be reportable", status)
		}
	}
	if ReportablePublicMedia("rejected") {
		t.Fatal("rejected media must not be reportable")
	}
}
