package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260928000006ExtendAnnouncementsForStatus struct{}

func (m *M20260928000006ExtendAnnouncementsForStatus) Signature() string {
	return "20260928000006_extend_announcements_for_status"
}

func (m *M20260928000006ExtendAnnouncementsForStatus) Up() error {
	if !facades.Schema().HasTable("announcements") {
		return nil
	}
	if !facades.Schema().HasColumn("announcements", "body") {
		if err := facades.Schema().Table("announcements", func(table schema.Blueprint) { table.Text("body").Nullable() }); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("announcements", "severity") {
		if err := facades.Schema().Table("announcements", func(table schema.Blueprint) { table.String("severity", 24).Default("info") }); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("announcements", "starts_at") {
		if err := facades.Schema().Table("announcements", func(table schema.Blueprint) { table.DateTimeTz("starts_at").Nullable() }); err != nil {
			return err
		}
	}
	if !facades.Schema().HasColumn("announcements", "ends_at") {
		if err := facades.Schema().Table("announcements", func(table schema.Blueprint) { table.DateTimeTz("ends_at").Nullable() }); err != nil {
			return err
		}
	}
	return nil
}

func (m *M20260928000006ExtendAnnouncementsForStatus) Down() error { return nil }
