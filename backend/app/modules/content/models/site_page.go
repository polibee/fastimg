package models

import "github.com/goravel/framework/database/orm"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

type SitePage struct {
	orm.Model
	Slug           string `json:"slug"`
	Title          string `json:"title"`
	ContentJSON    string `json:"content_json"`
	Excerpt        string `json:"excerpt"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	Status         string `json:"status"`
	PublishedAt    any    `json:"published_at"`
	CreatedBy      uint   `json:"created_by"`
	UpdatedBy      uint   `json:"updated_by"`
}
