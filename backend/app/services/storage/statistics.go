package storage

import (
	"time"

	"goravel/app/facades"
	"goravel/app/models"
)

// ConnectionStatistics is the site-side view of one enabled storage
// connection. It is intentionally based on FastImg's own object metadata and
// access records; provider billing/egress dashboards are not interchangeable
// with application usage and are not fabricated here.
type ConnectionStatistics struct {
	ConnectionID            uint              `json:"connection_id"`
	ProviderCode            string            `json:"provider_code"`
	Name                    string            `json:"name"`
	Status                  string            `json:"status"`
	IsPrimary               bool              `json:"is_primary"`
	ObjectCount             int64             `json:"object_count"`
	StoredBytes             int64             `json:"stored_bytes"`
	AllowedRequestCount     int64             `json:"allowed_request_count"`
	EstimatedBandwidthBytes int64             `json:"estimated_bandwidth_bytes"`
	MediaCount              int64             `json:"media_count"`
	ObjectStatusCounts      map[string]int64  `json:"object_status_counts"`
	VariantCounts           map[string]int64  `json:"variant_counts"`
	OrphanObjectCount       int64             `json:"orphan_object_count"`
	OrphanBytes             int64             `json:"orphan_bytes"`
	Daily                   []DailyStatistics `json:"daily"`
	Health                  HealthStatistics  `json:"health"`
	DataSource              string            `json:"data_source"`
	GeneratedAt             time.Time         `json:"generated_at"`
}

type DailyStatistics struct {
	Date                    string `json:"date"`
	UploadCount             int64  `json:"upload_count"`
	UploadBytes             int64  `json:"upload_bytes"`
	AllowedRequestCount     int64  `json:"allowed_request_count"`
	EstimatedBandwidthBytes int64  `json:"estimated_bandwidth_bytes"`
}

type HealthStatistics struct {
	LastStatus    string     `json:"last_status"`
	LastLatencyMS *int64     `json:"last_latency_ms,omitempty"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`
	ErrorCount24H int64      `json:"error_count_24h"`
}

type StatisticsOverview struct {
	GeneratedAt time.Time              `json:"generated_at"`
	Connections []ConnectionStatistics `json:"connections"`
}

type objectUsageRow struct {
	ObjectCount int64 `json:"object_count"`
	StoredBytes int64 `json:"stored_bytes"`
}

type accessUsageRow struct {
	AllowedRequestCount     int64 `json:"allowed_request_count"`
	EstimatedBandwidthBytes int64 `json:"estimated_bandwidth_bytes"`
}

type statusUsageRow struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type variantUsageRow struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type orphanUsageRow struct {
	Count int64 `json:"count"`
	Bytes int64 `json:"bytes"`
}

type dailyUploadRow struct {
	Date        string `json:"date"`
	UploadCount int64  `json:"upload_count"`
	UploadBytes int64  `json:"upload_bytes"`
}

type dailyAccessRow struct {
	Date                    string `json:"date"`
	AllowedRequestCount     int64  `json:"allowed_request_count"`
	EstimatedBandwidthBytes int64  `json:"estimated_bandwidth_bytes"`
}

type healthRow struct {
	Status    string    `json:"status"`
	LatencyMS *int64    `json:"latency_ms"`
	CheckedAt time.Time `json:"created_at"`
}

// Statistics returns live application-side figures for enabled connections.
// A missing optional table is treated as an empty metric so a fresh install
// can still open the page while migrations are being applied.
func (s *ConnectionService) Statistics() (StatisticsOverview, error) {
	now := time.Now().UTC()
	result := StatisticsOverview{GeneratedAt: now, Connections: make([]ConnectionStatistics, 0)}
	var connections []models.StorageConnection
	if err := facades.Orm().Query().Where("enabled = ?", true).OrderBy("is_primary", "desc").OrderBy("provider_code").Get(&connections); err != nil {
		return StatisticsOverview{}, err
	}
	for _, connection := range connections {
		item := ConnectionStatistics{
			ConnectionID:       connection.ID,
			ProviderCode:       connection.ProviderCode,
			Name:               connection.Name,
			Status:             connection.Status,
			IsPrimary:          connection.IsPrimary,
			DataSource:         "fastimg_database",
			GeneratedAt:        now,
			ObjectStatusCounts: make(map[string]int64),
			VariantCounts:      make(map[string]int64),
			Daily:              make([]DailyStatistics, 0),
		}
		if facades.Schema().HasTable("storage_objects") {
			var objects objectUsageRow
			if err := facades.Orm().Query().Raw(
				"SELECT COUNT(*) AS object_count, COALESCE(SUM(size_bytes), 0) AS stored_bytes FROM storage_objects WHERE storage_connection_id = ? AND status = ?",
				connection.ID, "ready",
			).Scan(&objects); err != nil {
				return StatisticsOverview{}, err
			}
			item.ObjectCount = objects.ObjectCount
			item.StoredBytes = objects.StoredBytes

			var statuses []statusUsageRow
			if err := facades.Orm().Query().Raw(
				"SELECT status, COUNT(*) AS count FROM storage_objects WHERE storage_connection_id = ? GROUP BY status ORDER BY status",
				connection.ID,
			).Scan(&statuses); err != nil {
				return StatisticsOverview{}, err
			}
			for _, status := range statuses {
				item.ObjectStatusCounts[status.Status] = status.Count
			}

			var orphan orphanUsageRow
			if facades.Schema().HasTable("media_variants") {
				if err := facades.Orm().Query().Raw(
					"SELECT COUNT(*) AS count, COALESCE(SUM(so.size_bytes), 0) AS bytes FROM storage_objects so WHERE so.storage_connection_id = ? AND NOT EXISTS (SELECT 1 FROM media_variants mv WHERE mv.storage_object_id = so.id)",
					connection.ID,
				).Scan(&orphan); err != nil {
					return StatisticsOverview{}, err
				}
			}
			item.OrphanObjectCount = orphan.Count
			item.OrphanBytes = orphan.Bytes
		}
		if facades.Schema().HasTable("media_access_logs") && facades.Schema().HasTable("media_variants") && facades.Schema().HasTable("storage_objects") {
			var accesses accessUsageRow
			if err := facades.Orm().Query().Raw(
				"SELECT COUNT(*) AS allowed_request_count, COALESCE(SUM(so.size_bytes), 0) AS estimated_bandwidth_bytes FROM media_access_logs mal INNER JOIN media_variants mv ON mv.media_asset_id = mal.media_asset_id AND mv.name = mal.variant INNER JOIN storage_objects so ON so.id = mv.storage_object_id WHERE so.storage_connection_id = ? AND mal.result = ?",
				connection.ID, "allowed",
			).Scan(&accesses); err != nil {
				return StatisticsOverview{}, err
			}
			item.AllowedRequestCount = accesses.AllowedRequestCount
			item.EstimatedBandwidthBytes = accesses.EstimatedBandwidthBytes

			var variants []variantUsageRow
			if err := facades.Orm().Query().Raw(
				"SELECT mv.name, COUNT(*) AS count FROM media_variants mv INNER JOIN storage_objects so ON so.id = mv.storage_object_id WHERE so.storage_connection_id = ? GROUP BY mv.name ORDER BY mv.name",
				connection.ID,
			).Scan(&variants); err != nil {
				return StatisticsOverview{}, err
			}
			for _, variant := range variants {
				item.VariantCounts[variant.Name] = variant.Count
			}

			var mediaCount struct {
				Count int64 `json:"count"`
			}
			if err := facades.Orm().Query().Raw(
				"SELECT COUNT(DISTINCT mv.media_asset_id) AS count FROM media_variants mv INNER JOIN storage_objects so ON so.id = mv.storage_object_id WHERE so.storage_connection_id = ?",
				connection.ID,
			).Scan(&mediaCount); err != nil {
				return StatisticsOverview{}, err
			}
			item.MediaCount = mediaCount.Count

			daily, err := dailyStatistics(connection.ID, now.AddDate(0, 0, -30))
			if err != nil {
				return StatisticsOverview{}, err
			}
			item.Daily = daily
		}
		health, err := healthStatistics(connection, now)
		if err != nil {
			return StatisticsOverview{}, err
		}
		item.Health = health
		result.Connections = append(result.Connections, item)
	}
	return result, nil
}

func dailyStatistics(connectionID uint, since time.Time) ([]DailyStatistics, error) {
	byDate := make(map[string]DailyStatistics)
	var uploads []dailyUploadRow
	if err := facades.Orm().Query().Raw(
		"SELECT TO_CHAR(DATE_TRUNC('day', ma.created_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS date, COUNT(DISTINCT ma.id) AS upload_count, COALESCE(SUM(CASE WHEN mv.name = 'original' THEN so.size_bytes ELSE 0 END), 0) AS upload_bytes FROM media_assets ma INNER JOIN media_variants mv ON mv.media_asset_id = ma.id INNER JOIN storage_objects so ON so.id = mv.storage_object_id WHERE so.storage_connection_id = ? AND ma.created_at >= ? AND ma.deleted_at IS NULL AND ma.status <> ? GROUP BY 1 ORDER BY 1",
		connectionID, since, "failed",
	).Scan(&uploads); err != nil {
		return nil, err
	}
	for _, upload := range uploads {
		byDate[upload.Date] = DailyStatistics{Date: upload.Date, UploadCount: upload.UploadCount, UploadBytes: upload.UploadBytes}
	}
	var accesses []dailyAccessRow
	if facades.Schema().HasTable("media_access_logs") {
		if err := facades.Orm().Query().Raw(
			"SELECT TO_CHAR(DATE_TRUNC('day', mal.accessed_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS date, COUNT(*) AS allowed_request_count, COALESCE(SUM(so.size_bytes), 0) AS estimated_bandwidth_bytes FROM media_access_logs mal INNER JOIN media_variants mv ON mv.media_asset_id = mal.media_asset_id AND mv.name = mal.variant INNER JOIN storage_objects so ON so.id = mv.storage_object_id WHERE so.storage_connection_id = ? AND mal.result = ? AND mal.accessed_at >= ? GROUP BY 1 ORDER BY 1",
			connectionID, "allowed", since,
		).Scan(&accesses); err != nil {
			return nil, err
		}
	}
	for _, access := range accesses {
		item := byDate[access.Date]
		item.Date = access.Date
		item.AllowedRequestCount = access.AllowedRequestCount
		item.EstimatedBandwidthBytes = access.EstimatedBandwidthBytes
		byDate[access.Date] = item
	}
	result := make([]DailyStatistics, 0, 30)
	for day := since.UTC().Truncate(24 * time.Hour); !day.After(time.Now().UTC()); day = day.Add(24 * time.Hour) {
		date := day.Format("2006-01-02")
		item := byDate[date]
		item.Date = date
		result = append(result, item)
	}
	return result, nil
}

func healthStatistics(connection models.StorageConnection, now time.Time) (HealthStatistics, error) {
	result := HealthStatistics{LastStatus: connection.Status, LastCheckedAt: connection.LastCheckedAt}
	if !facades.Schema().HasTable("storage_health_checks") {
		return result, nil
	}
	var latest healthRow
	if err := facades.Orm().Query().Raw(
		"SELECT status, latency_ms, created_at FROM storage_health_checks WHERE storage_connection_id = ? ORDER BY created_at DESC LIMIT 1",
		connection.ID,
	).Scan(&latest); err != nil {
		return HealthStatistics{}, err
	}
	if latest.Status != "" {
		result.LastStatus = latest.Status
		result.LastCheckedAt = &latest.CheckedAt
		result.LastLatencyMS = latest.LatencyMS
	}
	if err := facades.Orm().Query().Raw(
		"SELECT COUNT(*) AS count FROM storage_health_checks WHERE storage_connection_id = ? AND status = ? AND created_at >= ?",
		connection.ID, models.StorageConnectionError, now.Add(-24*time.Hour),
	).Scan(&result.ErrorCount24H); err != nil {
		return HealthStatistics{}, err
	}
	return result, nil
}
