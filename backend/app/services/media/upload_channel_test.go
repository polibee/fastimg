package media

import "testing"

func TestUploadChannelOnlyCountsPersonalAPIUploadsAgainstMonthlyAPILimit(t *testing.T) {
	if !CountsTowardMonthlyAPIUploads(UploadChannelAPI) {
		t.Fatal("personal API uploads must count toward the monthly API limit")
	}
	if CountsTowardMonthlyAPIUploads(UploadChannelMember) {
		t.Fatal("browser member uploads must not count toward the monthly API limit")
	}
}
