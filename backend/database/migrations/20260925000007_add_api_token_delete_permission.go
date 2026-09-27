package migrations

import (
	"goravel/app/facades"
	"goravel/app/models"
)

// M20260925000007AddAPITokenDeletePermission makes the destructive admin
// action explicit instead of treating view/update as permission to remove
// credentials.
type M20260925000007AddAPITokenDeletePermission struct{}

func (m *M20260925000007AddAPITokenDeletePermission) Signature() string {
	return "20260925000007_add_api_token_delete_permission"
}

func (m *M20260925000007AddAPITokenDeletePermission) Up() error {
	if !facades.Schema().HasTable("permissions") {
		return nil
	}
	permission := models.Permission{Name: "admin.api_tokens.delete", DisplayName: "api_tokens.delete"}
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
	if !facades.Schema().HasTable("permission_role") {
		return nil
	}
	var roles []models.Role
	if err := facades.Orm().Query().Model(&models.Role{}).Where("name = ?", "super-admin").Get(&roles); err != nil {
		return err
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
	return nil
}

func (m *M20260925000007AddAPITokenDeletePermission) Down() error {
	if !facades.Schema().HasTable("permissions") {
		return nil
	}
	var permissions []models.Permission
	if err := facades.Orm().Query().Model(&models.Permission{}).Where("name = ?", "admin.api_tokens.delete").Get(&permissions); err != nil {
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
	return nil
}
