package content

import (
	"strings"

	contentservices "goravel/app/modules/content/services"
)

type PageRequest struct {
	Slug           string         `json:"slug"`
	Title          string         `json:"title"`
	Content        map[string]any `json:"content_json"`
	Excerpt        string         `json:"excerpt"`
	SEOTitle       string         `json:"seo_title"`
	SEODescription string         `json:"seo_description"`
}

func (r PageRequest) Input(id, operatorID uint) contentservices.SaveDraftInput {
	return contentservices.SaveDraftInput{
		ID: id, Slug: strings.ToLower(strings.TrimSpace(r.Slug)), Title: strings.TrimSpace(r.Title),
		Content: r.Content, Excerpt: strings.TrimSpace(r.Excerpt), SEOTitle: strings.TrimSpace(r.SEOTitle),
		SEODescription: strings.TrimSpace(r.SEODescription), OperatorID: operatorID,
	}
}
