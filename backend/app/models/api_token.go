package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

// ApiToken stores only a digest of a Personal API Token. The raw value is
// intentionally not representable by this model so it cannot leak through
// generic JSON serialization or admin resources.
type ApiToken struct {
	orm.Model
	UserID      uint       `json:"user_id"`
	Name        string     `json:"name"`
	TokenHash   string     `json:"-"`
	TokenPrefix string     `json:"prefix"`
	ScopesJSON  string     `json:"-"`
	Status      string     `json:"status"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	LastUsedIP  string     `json:"last_used_ip"`
	UsageCount  int64      `json:"usage_count"`
	RevokedAt   *time.Time `json:"revoked_at"`
}
