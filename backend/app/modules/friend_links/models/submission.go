package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusHidden   = "hidden"
)

type Submission struct {
	orm.Model
	SiteName     string     `json:"site_name" gorm:"column:site_name"`
	URL          string     `json:"url" gorm:"column:url"`
	LogoURL      string     `json:"logo_url" gorm:"column:logo_url"`
	Description  string     `json:"description" gorm:"column:description"`
	ContactEmail string     `json:"contact_email" gorm:"column:contact_email"`
	SubmittedBy  *uint      `json:"submitted_by" gorm:"column:submitted_by"`
	Status       string     `json:"status" gorm:"column:status"`
	ReviewNote   string     `json:"review_note" gorm:"column:review_note"`
	ReviewedBy   *uint      `json:"reviewed_by" gorm:"column:reviewed_by"`
	ReviewedAt   *time.Time `json:"reviewed_at" gorm:"column:reviewed_at"`
}
