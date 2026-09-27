package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type EmailVerificationToken struct {
	orm.Model
	UserID     uint       `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
}
