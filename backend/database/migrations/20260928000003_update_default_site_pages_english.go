package migrations

import (
	"time"

	"goravel/app/facades"
	contentdefaults "goravel/app/modules/content/defaults"
)

// M20260928000003UpdateDefaultSitePagesEnglish upgrades the original Chinese
// seed content without overwriting a page that an administrator has edited.
// New installations receive the same English defaults from the seeder.
type M20260928000003UpdateDefaultSitePagesEnglish struct{}

func (m *M20260928000003UpdateDefaultSitePagesEnglish) Signature() string {
	return "20260928000003_update_default_site_pages_english"
}

func (m *M20260928000003UpdateDefaultSitePagesEnglish) Up() error {
	if !facades.Schema().HasTable("site_pages") {
		return nil
	}
	for _, page := range contentdefaults.DefaultPages() {
		_, err := facades.Orm().Query().Table("site_pages").
			Where("slug = ? AND title = ? AND excerpt = ?", page.Slug, page.OldTitle, page.OldExcerpt).
			Update(map[string]any{
				"title": page.Title, "excerpt": page.Excerpt, "seo_title": page.SEOTitle,
				"seo_description": page.SEODescription, "content_json": page.Content,
				"updated_at": time.Now().UTC(),
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *M20260928000003UpdateDefaultSitePagesEnglish) Down() error { return nil }
