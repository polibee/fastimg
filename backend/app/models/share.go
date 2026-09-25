package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type ShareLink struct {
	orm.Model
	UserID       uint       `json:"user_id"`
	MediaAssetID uint       `json:"media_asset_id"`
	TokenHash    string     `json:"-"`
	TokenPrefix  string     `json:"token_prefix"`
	PasswordHash string     `json:"-"`
	Status       string     `json:"status"`
	ExpiresAt    *time.Time `json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
}
