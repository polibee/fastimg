package models

import "github.com/goravel/framework/database/orm"

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusHidden   = "hidden"
)

type Submission struct {
	orm.Model
	SiteName     string `json:"site_name"`
	URL          string `json:"url"`
	LogoURL      string `json:"logo_url"`
	Description  string `json:"description"`
	ContactEmail string `json:"contact_email"`
	SubmittedBy  *uint  `json:"submitted_by"`
	Status       string `json:"status"`
	ReviewNote   string `json:"review_note"`
	ReviewedBy   *uint  `json:"reviewed_by"`
	ReviewedAt   any    `json:"reviewed_at"`
}
