package content

import (
	"testing"

	"github.com/goravel/framework/database/orm"

	contentmodels "goravel/app/modules/content/models"
)

func TestPublicPagePayloadExcludesAdministrativeFields(t *testing.T) {
	page := contentmodels.SitePage{
		Model: orm.Model{ID: 7}, Slug: "privacy", Title: "Privacy", ContentJSON: `{"type":"doc"}`,
		Status: contentmodels.StatusPublished, CreatedBy: 12, UpdatedBy: 13,
	}

	payload := PublicPagePayload(page)
	if payload["slug"] != "privacy" || payload["title"] != "Privacy" {
		t.Fatalf("public payload lost page fields: %#v", payload)
	}
	for _, key := range []string{"id", "status", "created_by", "updated_by", "content_json"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("public payload exposed administrative field %q: %#v", key, payload)
		}
	}
}

func TestPageSlugAndStatusRules(t *testing.T) {
	for _, slug := range []string{"privacy", "about-us", "page-2"} {
		if !ValidPageSlug(slug) {
			t.Errorf("ValidPageSlug(%q) = false", slug)
		}
	}
	for _, slug := range []string{"Privacy", "../admin", "two words", ""} {
		if ValidPageSlug(slug) {
			t.Errorf("ValidPageSlug(%q) = true", slug)
		}
	}
	if !CanPublish(contentmodels.StatusDraft) || !CanPublish(contentmodels.StatusArchived) {
		t.Fatal("draft and archived pages should be publishable")
	}
	if CanPublish(contentmodels.StatusPublished) {
		t.Fatal("published page should not need a publish transition")
	}
}
