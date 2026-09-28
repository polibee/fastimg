package seeders

import (
	"encoding/json"
	"time"

	"goravel/app/facades"
	contentmodels "goravel/app/modules/content/models"
)

type SiteContent struct{}

func (s *SiteContent) Signature() string { return "SiteContent" }

func (s *SiteContent) Run() error {
	if !facades.Schema().HasTable("site_pages") {
		return nil
	}
	pages := []struct {
		slug, title, excerpt, body string
	}{
		{"privacy", "隐私政策", "FastImg 如何处理账户、媒体和访问数据。", "FastImg 只收集提供媒体托管、稳定链接、账户安全和服务支持所必需的信息。你上传的媒体归属于你的账户；访问记录用于安全、用量统计和故障排查。"},
		{"terms", "使用条款", "使用 FastImg 服务时需要遵守的基本规则。", "请仅上传你拥有合法权利或获得授权的内容。不得使用 FastImg 托管违法、侵权、恶意软件或绕过安全控制的内容。违规内容可能被隐藏、删除，相关账户可能受到限制。"},
		{"about", "关于我们", "FastImg 面向开发者、站长和内容创作者。", "FastImg 提供稳定图片链接、媒体管理、API 上传、访问控制和会员套餐，帮助你更简单地管理和分享媒体资产。"},
	}
	for _, item := range pages {
		content, err := paragraphDocument(item.body)
		if err != nil {
			return err
		}
		var existing []contentmodels.SitePage
		if err := facades.Orm().Query().Where("slug = ?", item.slug).Get(&existing); err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		now := time.Now().UTC()
		if err := facades.Orm().Query().Table("site_pages").Create(&map[string]any{
			"slug": item.slug, "title": item.title, "excerpt": item.excerpt, "content_json": content,
			"status": contentmodels.StatusPublished, "published_at": now, "created_at": now, "updated_at": now,
		}); err != nil {
			return err
		}
	}
	return ensureDefaultFooterNavigation()
}

func paragraphDocument(text string) (string, error) {
	return marshalJSON(map[string]any{
		"type": "doc",
		"content": []any{map[string]any{
			"type":    "paragraph",
			"content": []any{map[string]any{"type": "text", "text": text}},
		}},
	})
}

func marshalJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func ensureDefaultFooterNavigation() error {
	if !facades.Schema().HasTable("footer_navigation_groups") || !facades.Schema().HasTable("footer_navigation_items") {
		return nil
	}
	groups := []struct {
		title string
		order int
	}{
		{"法律与隐私", 10},
		{"关于", 20},
	}
	for _, group := range groups {
		var existing []map[string]any
		if err := facades.Orm().Query().Table("footer_navigation_groups").Where("title = ? AND locale = ?", group.title, "all").Get(&existing); err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		if err := facades.Orm().Query().Table("footer_navigation_groups").Create(&map[string]any{
			"title": group.title, "locale": "all", "sort_order": group.order, "is_enabled": true,
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
			label, targetType, targetValue string
			order                          int
		}{
			{"隐私政策", "page", "privacy", 10},
			{"使用条款", "page", "terms", 20},
		}
		if group["title"] == "关于" {
			items = []struct {
				label, targetType, targetValue string
				order                          int
			}{
				{"关于我们", "page", "about", 10},
				{"友情链接", "friends", "friends", 20},
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
				"group_id": groupID, "label": item.label, "target_type": item.targetType, "target_value": item.targetValue,
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
