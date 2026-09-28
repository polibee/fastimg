package migrations

import (
	"goravel/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260928000001CreateContentNavigationFriendLinksTables struct{}

func (m *M20260928000001CreateContentNavigationFriendLinksTables) Signature() string {
	return "20260928000001_create_content_navigation_friend_links_tables"
}

func (m *M20260928000001CreateContentNavigationFriendLinksTables) Up() error {
	if !facades.Schema().HasTable("site_pages") {
		if err := facades.Schema().Create("site_pages", func(table schema.Blueprint) {
			table.ID()
			table.String("slug", 160)
			table.String("title", 255)
			table.Text("content_json")
			table.Text("excerpt").Nullable()
			table.String("seo_title", 255).Nullable()
			table.Text("seo_description").Nullable()
			table.String("status", 24).Default("draft")
			table.DateTimeTz("published_at").Nullable()
			table.UnsignedBigInteger("created_by").Nullable()
			table.UnsignedBigInteger("updated_by").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Unique("slug")
			table.Index("status")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("footer_navigation_groups") {
		if err := facades.Schema().Create("footer_navigation_groups", func(table schema.Blueprint) {
			table.ID()
			table.String("title", 160)
			table.String("locale", 16).Default("all")
			table.Integer("sort_order").Default(0)
			table.Boolean("is_enabled").Default(true)
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("locale", "is_enabled", "sort_order")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("footer_navigation_items") {
		if err := facades.Schema().Create("footer_navigation_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("group_id")
			table.UnsignedBigInteger("parent_id").Nullable()
			table.String("label", 160)
			table.String("target_type", 24)
			table.String("target_value", 512)
			table.Boolean("open_in_new_tab").Default(false)
			table.Integer("sort_order").Default(0)
			table.Boolean("is_enabled").Default(true)
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("group_id", "parent_id", "sort_order")
			table.Index("is_enabled")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("friend_link_submissions") {
		return facades.Schema().Create("friend_link_submissions", func(table schema.Blueprint) {
			table.ID()
			table.String("site_name", 160)
			table.String("url", 2048)
			table.String("logo_url", 2048).Nullable()
			table.Text("description").Nullable()
			table.String("contact_email", 320).Nullable()
			table.UnsignedBigInteger("submitted_by").Nullable()
			table.String("status", 24).Default("pending")
			table.Text("review_note").Nullable()
			table.UnsignedBigInteger("reviewed_by").Nullable()
			table.DateTimeTz("reviewed_at").Nullable()
			table.DateTimeTz("created_at").Nullable()
			table.DateTimeTz("updated_at").Nullable()
			table.Index("status", "created_at")
			table.Index("url")
			table.Index("contact_email", "created_at")
		})
	}
	return nil
}

func (m *M20260928000001CreateContentNavigationFriendLinksTables) Down() error {
	for _, table := range []string{"friend_link_submissions", "footer_navigation_items", "footer_navigation_groups", "site_pages"} {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}
	return nil
}
