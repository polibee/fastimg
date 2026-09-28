package seeders

import (
	"time"

	"goravel/app/facades"
	contentdefaults "goravel/app/modules/content/defaults"
	contentmodels "goravel/app/modules/content/models"
)

type SiteContent struct{}

func (s *SiteContent) Signature() string { return "SiteContent" }

func (s *SiteContent) Run() error {
	if !facades.Schema().HasTable("site_pages") {
		return nil
	}
	for _, item := range contentdefaults.DefaultPages() {
		var existing []map[string]any
		if err := facades.Orm().Query().Where("slug = ?", item.Slug).Get(&existing); err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		now := time.Now().UTC()
		if err := facades.Orm().Query().Table("site_pages").Create(&map[string]any{
			"slug": item.Slug, "title": item.Title, "excerpt": item.Excerpt, "seo_title": item.SEOTitle, "seo_description": item.SEODescription, "content_json": item.Content,
			"status": contentmodels.StatusPublished, "published_at": now, "created_at": now, "updated_at": now,
		}); err != nil {
			return err
		}
	}
	return ensureDefaultFooterNavigation()
}

func ensureDefaultFooterNavigation() error {
	if !facades.Schema().HasTable("footer_navigation_groups") || !facades.Schema().HasTable("footer_navigation_items") {
		return nil
	}
	groups := []struct {
		titleZhCN, titleEnUS string
		order                int
	}{
		{"法律与隐私", "Legal & privacy", 10},
		{"关于", "About", 20},
	}
	for _, group := range groups {
		var existing []map[string]any
		if err := facades.Orm().Query().Table("footer_navigation_groups").Where("title = ? AND locale = ?", group.titleZhCN, "all").Get(&existing); err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		if err := facades.Orm().Query().Table("footer_navigation_groups").Create(&map[string]any{
			"title": group.titleZhCN, "title_zh_cn": group.titleZhCN, "title_en_us": group.titleEnUS, "locale": "all", "sort_order": group.order, "is_enabled": true,
			"created_at": time.Now().UTC(), "updated_at": time.Now().UTC(),
		}); err != nil {
			return err
		}
	}
	return ensureDefaultFooterItems()
}

func ensureDefaultFooterItems() error {
	var groups []map[string]any
	if err := facades.Orm().Query().Table("footer_navigation_groups").Where("title IN (?, ?)", "法律与隐私", "关于").Get(&groups); err != nil {
		return err
	}
	for _, group := range groups {
		groupID, ok := numericID(group["id"])
		if !ok {
			continue
		}
		items := []struct {
			labelZhCN, labelEnUS, targetType, targetValue string
			order                                         int
		}{
			{"隐私政策", "Privacy policy", "page", "privacy", 10},
			{"使用条款", "Terms of use", "page", "terms", 20},
		}
		if group["title"] == "关于" {
			items = []struct {
				labelZhCN, labelEnUS, targetType, targetValue string
				order                                         int
			}{
				{"关于我们", "About us", "page", "about", 10},
				{"友情链接", "Friend links", "friends", "friends", 20},
			}
		}
		for _, item := range items {
			var existing []map[string]any
			if err := facades.Orm().Query().Table("footer_navigation_items").Where("group_id = ? AND target_type = ? AND target_value = ?", groupID, item.targetType, item.targetValue).Get(&existing); err != nil {
				return err
			}
			if len(existing) > 0 {
				continue
			}
			if err := facades.Orm().Query().Table("footer_navigation_items").Create(&map[string]any{
				"group_id": groupID, "label": item.labelZhCN, "label_zh_cn": item.labelZhCN, "label_en_us": item.labelEnUS, "target_type": item.targetType, "target_value": item.targetValue,
				"sort_order": item.order, "is_enabled": true, "created_at": time.Now().UTC(), "updated_at": time.Now().UTC(),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func numericID(value any) (uint, bool) {
	switch number := value.(type) {
	case uint:
		return number, number > 0
	case uint64:
		return uint(number), number > 0
	case int:
		return uint(number), number > 0
	case int64:
		return uint(number), number > 0
	default:
		return 0, false
	}
}
