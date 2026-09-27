package statistics

import (
	"testing"
	"time"
)

func TestBuildTrendPointsCreatesContinuousRangeAndBucketsRealEvents(t *testing.T) {
	from := time.Date(2026, 9, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	to := from.AddDate(0, 0, 2)
	points, err := BuildTrendPoints([]TrendEvent{
		{Kind: "users", At: time.Date(2026, 9, 1, 1, 0, 0, 0, time.UTC)},
		{Kind: "media", At: time.Date(2026, 9, 2, 4, 0, 0, 0, time.UTC)},
		{Kind: "bandwidth", At: time.Date(2026, 9, 2, 5, 0, 0, 0, time.UTC), Value: 2048},
		{Kind: "orders", At: time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)},
	}, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 3 {
		t.Fatalf("points = %d, want 3", len(points))
	}
	if points[0].Users != 1 || points[1].Media != 1 || points[1].BandwidthBytes != 2048 || points[2].Users != 0 {
		t.Fatalf("unexpected points: %#v", points)
	}
}

func TestBuildTrendPointsRejectsUnboundedRange(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := BuildTrendPoints(nil, from, from.AddDate(0, 0, maxTrendDays+1)); err == nil {
		t.Fatal("expected range validation error")
	}
}
