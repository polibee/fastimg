package migrations

import (
	"goravel/app/facades"
	"goravel/app/models"
)

type M20260927000002AddStoragePermissions struct{}

func (m *M20260927000002AddStoragePermissions) Signature() string {
	return "20260927000002_add_storage_permissions"
}

func (m *M20260927000002AddStoragePermissions) Up() error {
	if !facades.Schema().HasTable("permissions") || !facades.Schema().HasTable("permission_role") {
		return nil
	}
	permissions := []models.Permission{
		{Name: "admin.storage.view", DisplayName: "storage.view"},
		{Name: "admin.storage.manage", DisplayName: "storage.manage"},
	}
	roles := make([]models.Role, 0)
	if err := facades.Orm().Query().Model(&models.Role{}).Where("name = ?", "super-admin").Get(&roles); err != nil {
		return err
	}
	for _, permission := range permissions {
		var existing []models.Permission
		if err := facades.Orm().Query().Model(&models.Permission{}).Where("name = ?", permission.Name).Get(&existing); err != nil {
			return err
		}
		if len(existing) == 0 {
			if err := facades.Orm().Query().Create(&permission); err != nil {
				return err
			}
			existing = []models.Permission{permission}
		}
		for _, role := range roles {
			assigned, err := facades.Orm().Query().Table("permission_role").Where("permission_id = ? AND role_id = ?", existing[0].ID, role.ID).Exists()
			if err != nil {
				return err
			}
			if !assigned {
				if err := facades.Orm().Query().Table("permission_role").Create(&map[string]any{"permission_id": existing[0].ID, "role_id": role.ID}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *M20260927000002AddStoragePermissions) Down() error {
	if !facades.Schema().HasTable("permissions") {
		return nil
	}
	for _, name := range []string{"admin.storage.view", "admin.storage.manage"} {
		var permissions []models.Permission
		if err := facades.Orm().Query().Model(&models.Permission{}).Where("name = ?", name).Get(&permissions); err != nil {
			return err
		}
		for _, permission := range permissions {
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
