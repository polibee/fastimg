package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260920000001CreateUsersTable{},
		&migrations.M20260920000002CreateRBACTables{},
		&migrations.M20260920000003AddRBACTimestamps{},
		&migrations.M20260921000001ReplaceUserActiveWithStatus{},
		&migrations.M20260921000002CreateAuthRefreshTokensTable{},
		&migrations.M20260921000003CreateAuditLogsTable{},
		&migrations.M20260921000004CreateAuthLoginAttemptsTable{},
		&migrations.MannouncementsCreateAnnouncementsTable{},
		&migrations.M20260922000001AddPermissionRoleScope{},
		&migrations.M20260922000002CreatePermissionRoleFieldsTable{},
		&migrations.MdepartmentsCreateDepartmentsTable{},
		&migrations.M20260922000004AddDepartmentsParentID{},
		&migrations.M20260922000005CreateNotificationsTable{},
		&migrations.M20260923000001CreatePlansSubscriptionsUsageTables{},
		&migrations.M20260923000002CreateFastimgMediaTables{},
		&migrations.MadvertisingCreateAdvertisingTable{},
		&migrations.MfoldersCreateFoldersTable{},
		&migrations.MalbumsCreateAlbumsTable{},
		&migrations.M20260924000001AddMediaFolderID{},
		&migrations.M20260924000002CreateShareLinksTable{},
		&migrations.M20260924000003CreateAPITokensTable{},
		&migrations.M20260924000004AddAdvertisingCreativeFields{},
		&migrations.M20260924000005CreateLinkSecurityTables{},
		&migrations.M20260924000006CreateBillingTables{},
	}
}
