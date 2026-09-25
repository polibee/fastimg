package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20260924000004AddAdvertisingCreativeFields keeps legacy creative_url rows
// readable while giving administrators explicit text, image, and script content.
type M20260924000004AddAdvertisingCreativeFields struct{}

func (m *M20260924000004AddAdvertisingCreativeFields) Signature() string {
	return "20260924000004_add_advertising_creative_fields"
}

func (m *M20260924000004AddAdvertisingCreativeFields) Up() error {
	if !facades.Schema().HasTable("advertising") {
		return nil
	}
	return facades.Schema().Table("advertising", func(table schema.Blueprint) {
		if !facades.Schema().HasColumn("advertising", "creative_type") {
			table.String("creative_type").Default("image")
		}
		if !facades.Schema().HasColumn("advertising", "creative_content") {
			table.Text("creative_content").Nullable()
		}
	})
}

func (m *M20260924000004AddAdvertisingCreativeFields) Down() error {
	if !facades.Schema().HasTable("advertising") {
		return nil
	}
	return facades.Schema().Table("advertising", func(table schema.Blueprint) {
		if facades.Schema().HasColumn("advertising", "creative_content") {
			table.DropColumn("creative_content")
		}
		if facades.Schema().HasColumn("advertising", "creative_type") {
			table.DropColumn("creative_type")
		}
	})
}
