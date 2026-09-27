package statistics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"goravel/app/facades"
)

const maxTrendDays = 93

type TrendEvent struct {
	Kind  string
	At    time.Time
	Value int64
}

type TrendPoint struct {
	Date                string `json:"date"`
	Users               int64  `json:"users"`
	Media               int64  `json:"media"`
	Albums              int64  `json:"albums"`
	BandwidthBytes      int64  `json:"bandwidth_bytes"`
	Orders              int64  `json:"orders"`
	PaymentTransactions int64  `json:"payment_transactions"`
}

type Report struct {
	From   string       `json:"from"`
	To     string       `json:"to"`
	Points []TrendPoint `json:"points"`
}

type Service struct{}

func NewService() *Service { return &Service{} }

// BuildTrendPoints keeps date bucketing deterministic and database-agnostic.
// The database adapter only supplies timestamped events; aggregation remains
// here so PostgreSQL/MySQL differences do not leak into the admin contract.
func BuildTrendPoints(events []TrendEvent, from, to time.Time) ([]TrendPoint, error) {
	from = startOfDay(from)
	to = startOfDay(to)
	if from.IsZero() || to.Before(from) || to.Sub(from) > maxTrendDays*24*time.Hour {
		return nil, fmt.Errorf("trend range must be between 1 and %d days", maxTrendDays)
	}
	points := make([]TrendPoint, 0, int(to.Sub(from)/(24*time.Hour))+1)
	index := make(map[string]int)
	for day := from; !day.After(to); day = day.Add(24 * time.Hour) {
		date := day.Format("2006-01-02")
		index[date] = len(points)
		points = append(points, TrendPoint{Date: date})
	}
	for _, event := range events {
		at := event.At.UTC()
		if at.Before(from) || at.After(to.Add(24*time.Hour)) {
			continue
		}
		position, ok := index[at.Format("2006-01-02")]
		if !ok {
			continue
		}
		switch event.Kind {
		case "users":
			points[position].Users++
		case "media":
			points[position].Media++
		case "albums":
			points[position].Albums++
		case "bandwidth":
			points[position].BandwidthBytes += event.Value
		case "orders":
			points[position].Orders++
		case "payment_transactions":
			points[position].PaymentTransactions++
		}
	}
	return points, nil
}

func (s *Service) Trends(ctx context.Context, from, to time.Time) (Report, error) {
	points, err := BuildTrendPoints(nil, from, to)
	if err != nil {
		return Report{}, err
	}
	from = startOfDay(from)
	toExclusive := startOfDay(to).Add(24 * time.Hour)
	events := make([]TrendEvent, 0)
	loaders := []func() error{
		func() error {
			return loadTimestampEvents(ctx, "users", "created_at", "users", from, toExclusive, &events)
		},
		func() error {
			return loadTimestampEvents(ctx, "media_assets", "created_at", "media", from, toExclusive, &events)
		},
		func() error {
			return loadTimestampEvents(ctx, "albums", "created_at", "albums", from, toExclusive, &events)
		},
		func() error {
			return loadTimestampEvents(ctx, "orders", "created_at", "orders", from, toExclusive, &events)
		},
		func() error {
			return loadTimestampEvents(ctx, "payment_transactions", "occurred_at", "payment_transactions", from, toExclusive, &events)
		},
		func() error { return loadBandwidthEvents(ctx, from, toExclusive, &events) },
	}
	for _, load := range loaders {
		if err := load(); err != nil {
			return Report{}, fmt.Errorf("load trend events: %w", err)
		}
	}
	points, err = BuildTrendPoints(events, from, to)
	if err != nil {
		return Report{}, err
	}
	return Report{From: from.Format("2006-01-02"), To: startOfDay(to).Format("2006-01-02"), Points: points}, nil
}

type timestampRow struct {
	At time.Time `json:"at"`
}
type bandwidthRow struct {
	At    time.Time `json:"at"`
	Value int64     `json:"value"`
}

func loadTimestampEvents(_ context.Context, table, column, kind string, from, to time.Time, events *[]TrendEvent) error {
	if !facades.Schema().HasTable(table) {
		return nil
	}
	var rows []timestampRow
	if err := facades.Orm().Query().Table(table).Select(column+" as at").Where(column+" >= ? AND "+column+" < ?", from, to).Get(&rows); err != nil {
		return fmt.Errorf("query %s.%s: %w", table, column, err)
	}
	for _, row := range rows {
		*events = append(*events, TrendEvent{Kind: kind, At: row.At})
	}
	return nil
}

func loadBandwidthEvents(_ context.Context, from, to time.Time, events *[]TrendEvent) error {
	if !facades.Schema().HasTable("usage_ledgers") {
		return nil
	}
	var rows []bandwidthRow
	if err := facades.Orm().Query().Table("usage_ledgers").Select("created_at as at", "delta as value").Where("resource_type = ? AND source_type = ? AND created_at >= ? AND created_at < ?", "bandwidth", "download", from, to).Get(&rows); err != nil {
		return fmt.Errorf("query usage_ledgers bandwidth events: %w", err)
	}
	for _, row := range rows {
		*events = append(*events, TrendEvent{Kind: "bandwidth", At: row.At, Value: row.Value})
	}
	return nil
}

func startOfDay(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func SortEvents(events []TrendEvent) {
	sort.SliceStable(events, func(i, j int) bool { return events[i].At.Before(events[j].At) })
}
