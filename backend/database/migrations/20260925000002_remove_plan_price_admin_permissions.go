package migrations

import (
	"goravel/app/facades"
	"goravel/app/models"
)

// M20260925000002RemovePlanPriceAdminPermissions removes the temporary
// standalone admin resource for plan_prices. The price catalog table remains
// a billing-internal table used by checkout and historical order snapshots.
type M20260925000002RemovePlanPriceAdminPermissions struct{}

func (m *M20260925000002RemovePlanPriceAdminPermissions) Signature() string {
	return "20260925000002_remove_plan_price_admin_permissions"
}

func (m *M20260925000002RemovePlanPriceAdminPermissions) Up() error {
	if !facades.Schema().HasTable("permissions") {
		return nil
	}

	for _, name := range []string{
		"admin.plan_prices.view",
		"admin.plan_prices.create",
		"admin.plan_prices.update",
		"admin.plan_prices.delete",
	} {
		var permissions []models.Permission
		if err := facades.Orm().Query().Model(&models.Permission{}).Where("name = ?", name).Get(&permissions); err != nil {
			return err
		}
		for _, permission := range permissions {
			if facades.Schema().HasTable("permission_role_field") {
				if _, err := facades.Orm().Query().Table("permission_role_field").Where("permission_id = ?", permission.ID).Delete(); err != nil {
					return err
				}
			}
			if facades.Schema().HasTable("permission_role") {
				if _, err := facades.Orm().Query().Table("permission_role").Where("permission_id = ?", permission.ID).Delete(); err != nil {
					return err
				}
			}
			if _, err := facades.Orm().Query().Model(&models.Permission{}).Where("id = ?", permission.ID).Delete(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *M20260925000002RemovePlanPriceAdminPermissions) Down() error {
	// The standalone resource and permissions are intentionally removed during
	// development. Reintroducing them would recreate the duplicate admin
	// surface and is outside this migration's rollback contract.
	return nil
}
