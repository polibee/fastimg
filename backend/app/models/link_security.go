package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

// HotlinkPolicy stores the delivery policy for one owned media asset.
// A missing row is intentionally treated as the default off policy.
type HotlinkPolicy struct {
	orm.Model
	UserID         uint   `json:"user_id"`
	MediaAssetID   uint   `json:"media_asset_id"`
	Mode           string `json:"mode"`
	AllowNoReferer bool   `json:"allow_no_referer"`
}

type HotlinkDomain struct {
	orm.Model
	UserID uint   `json:"user_id"`
	Host   string `json:"host"`
	Status string `json:"status"`
}

// MediaAccessLog is intentionally append-only from the application service.
// It records the decision without storing query signatures or passwords.
type MediaAccessLog struct {
	orm.Model
	MediaAssetID uint      `json:"media_asset_id"`
	ShareLinkID  *uint     `json:"share_link_id"`
	Variant      string    `json:"variant"`
	DeliveryMode string    `json:"delivery_mode"`
	Result       string    `json:"result"`
	RefererHost  string    `json:"referer_host"`
	UserAgent    string    `json:"user_agent"`
	AccessedAt   time.Time `json:"accessed_at"`
}
