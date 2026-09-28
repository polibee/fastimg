package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

const (
	SecurityScanQueued   = "queued"
	SecurityScanRunning  = "running"
	SecurityScanClear    = "clear"
	SecurityScanHighRisk = "high_risk"
	SecurityScanFailed   = "failed"
)

type MediaSecurityScan struct {
	orm.Model
	MediaAssetID uint       `json:"media_asset_id"`
	Status       string     `json:"status"`
	Provider     string     `json:"provider"`
	Reason       string     `json:"reason"`
	FindingsJSON string     `json:"findings_json"`
	RiskScore    int        `json:"risk_score"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}
