package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type User struct {
	orm.Model
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Status   string `json:"status"`
	Locale   string `json:"locale"`
	// Stored separately in auth helpers so older feature-test schemas and
	// installations can boot before the verification migration is applied.
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" gorm:"-"`
}

func (u User) Public() map[string]any {
	return map[string]any{
		"id":                u.ID,
		"name":              u.Name,
		"email":             u.Email,
		"status":            u.Status,
		"locale":            u.Locale,
		"email_verified_at": u.EmailVerifiedAt,
		"created_at":        u.CreatedAt,
		"updated_at":        u.UpdatedAt,
	}
}
