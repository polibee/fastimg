package migrations

import (
	"goravel/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260928000002AddFooterNavigationBilingualLabels struct{}

func (m *M20260928000002AddFooterNavigationBilingualLabels) Signature() string {
	return "20260928000002_add_footer_navigation_bilingual_labels"
}

func (m *M20260928000002AddFooterNavigationBilingualLabels) Up() error {
	if facades.Schema().HasTable("footer_navigation_groups") {
		if !facades.Schema().HasColumn("footer_navigation_groups", "title_zh_cn") {
			if err := facades.Schema().Table("footer_navigation_groups", func(table schema.Blueprint) { table.String("title_zh_cn", 160).Nullable() }); err != nil {
				return err
			}
		}
		if !facades.Schema().HasColumn("footer_navigation_groups", "title_en_us") {
			if err := facades.Schema().Table("footer_navigation_groups", func(table schema.Blueprint) { table.String("title_en_us", 160).Nullable() }); err != nil {
				return err
			}
		}
		if _, err := facades.Orm().Query().Table("footer_navigation_groups").Where("title = ?", "法律与隐私").Update(map[string]any{"title_zh_cn": "法律与隐私", "title_en_us": "Legal & privacy"}); err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Table("footer_navigation_groups").Where("title = ?", "关于").Update(map[string]any{"title_zh_cn": "关于", "title_en_us": "About"}); err != nil {
			return err
		}
	}
	if facades.Schema().HasTable("footer_navigation_items") {
		if !facades.Schema().HasColumn("footer_navigation_items", "label_zh_cn") {
			if err := facades.Schema().Table("footer_navigation_items", func(table schema.Blueprint) { table.String("label_zh_cn", 160).Nullable() }); err != nil {
				return err
			}
		}
		if !facades.Schema().HasColumn("footer_navigation_items", "label_en_us") {
			if err := facades.Schema().Table("footer_navigation_items", func(table schema.Blueprint) { table.String("label_en_us", 160).Nullable() }); err != nil {
				return err
			}
		}
		known := map[string]string{"隐私政策": "Privacy policy", "使用条款": "Terms of use", "关于我们": "About us", "友情链接": "Friend links"}
		for zh, en := range known {
			if _, err := facades.Orm().Query().Table("footer_navigation_items").Where("label = ?", zh).Update(map[string]any{"label_zh_cn": zh, "label_en_us": en}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *M20260928000002AddFooterNavigationBilingualLabels) Down() error {
	if facades.Schema().HasTable("footer_navigation_items") {
		if facades.Schema().HasColumn("footer_navigation_items", "label_en_us") {
			if err := facades.Schema().Table("footer_navigation_items", func(table schema.Blueprint) { table.DropColumn("label_en_us") }); err != nil {
				return err
			}
		}
		if facades.Schema().HasColumn("footer_navigation_items", "label_zh_cn") {
			if err := facades.Schema().Table("footer_navigation_items", func(table schema.Blueprint) { table.DropColumn("label_zh_cn") }); err != nil {
				return err
			}
		}
	}
	if facades.Schema().HasTable("footer_navigation_groups") {
		if facades.Schema().HasColumn("footer_navigation_groups", "title_en_us") {
			if err := facades.Schema().Table("footer_navigation_groups", func(table schema.Blueprint) { table.DropColumn("title_en_us") }); err != nil {
				return err
			}
		}
		if facades.Schema().HasColumn("footer_navigation_groups", "title_zh_cn") {
			if err := facades.Schema().Table("footer_navigation_groups", func(table schema.Blueprint) { table.DropColumn("title_zh_cn") }); err != nil {
				return err
			}
		}
	}
	return nil
}
