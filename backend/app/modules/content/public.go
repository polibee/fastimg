package content

import (
	"encoding/json"
	"regexp"

	contentmodels "goravel/app/modules/content/models"
)

var publicPageSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidPageSlug(slug string) bool { return publicPageSlugPattern.MatchString(slug) }

func CanPublish(status string) bool {
	return status == contentmodels.StatusDraft || status == contentmodels.StatusArchived
}

// PublicPagePayload deliberately does not expose IDs, status, operator IDs or
// the raw persisted JSON string. The parsed document has already passed the
// content sanitizer before it is persisted.
func PublicPagePayload(page contentmodels.SitePage) map[string]any {
	payload := map[string]any{
		"slug": page.Slug, "title": page.Title, "excerpt": page.Excerpt,
		"seo_title": page.SEOTitle, "seo_description": page.SEODescription,
	}
	var document any
	if err := json.Unmarshal([]byte(page.ContentJSON), &document); err == nil {
		payload["content"] = document
	}
	return payload
}
