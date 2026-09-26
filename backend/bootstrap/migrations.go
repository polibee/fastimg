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
		&migrations.M20260921000005CreateSystemSettingsTable{},
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
		&migrations.M20260924000007AddPlanPriceGatewayCode{},
		&migrations.M20260924000008DropPlanPriceGatewayCode{},
		&migrations.M20260925000001CreateMediaReportsTable{},
		&migrations.M20260925000002RemovePlanPriceAdminPermissions{},
		&migrations.M20260925000003CreateAPITokenRateLimitsTable{},
		&migrations.M20260925000004CreateAlbumMediaTable{},
		&migrations.M20260925000005AddDiscoveryStateToMedia{},
		&migrations.M20260925000006AddWatermarkEntitlementDefaults{},
		&migrations.M20260925000007AddAPITokenDeletePermission{},
		&migrations.M20260926000001DefaultUploadedMediaApproved{},
		&migrations.M20260926000002AddEmailVerifiedAtToUsers{},
		&migrations.M20260926000003CreateEmailVerificationTokensTable{},
		&migrations.M20260926000004CreateBackupJobsTable{},
	}
}
