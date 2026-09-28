package console

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/facades"
	"goravel/app/models"
	userservices "goravel/app/services/users"
)

// AdminBootstrapCommand creates or promotes one explicitly supplied account.
// It intentionally reads the password from stdin so deployment scripts never
// put the credential in a process argument or a persisted environment file.
type AdminBootstrapCommand struct{}

func (AdminBootstrapCommand) Signature() string { return "admin:bootstrap" }

func (AdminBootstrapCommand) Description() string {
	return "Create or promote the first FastImg administrator from explicit input"
}

func (AdminBootstrapCommand) Extend() command.Extend {
	return command.Extend{
		Category: "admin",
		Flags: []command.Flag{
			&command.StringFlag{Name: "email", Usage: "administrator email address"},
			&command.StringFlag{Name: "name", Usage: "administrator display name", Value: "Administrator"},
		},
	}
}

func (AdminBootstrapCommand) Handle(ctx console.Context) error {
	email := strings.ToLower(strings.TrimSpace(ctx.Option("email")))
	name := strings.TrimSpace(ctx.Option("name"))
	if email == "" || !strings.Contains(email, "@") {
		return errors.New("--email is required and must be a valid email address")
	}
	if name == "" {
		name = "Administrator"
	}

	password, err := readBootstrapPassword()
	if err != nil {
		return err
	}
	if len([]rune(password)) < 12 {
		return errors.New("bootstrap password must contain at least 12 characters")
	}

	user, created, err := ensureBootstrapUser(name, email, password)
	if err != nil {
		return err
	}
	if err := ensureBootstrapRole(user.ID); err != nil {
		return err
	}

	if created {
		ctx.Success(fmt.Sprintf("Administrator created: %s", email))
	} else {
		ctx.Success(fmt.Sprintf("Administrator access ensured for existing account: %s (password unchanged)", email))
	}
	return nil
}

func readBootstrapPassword() (string, error) {
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", errors.New("administrator password must be provided on stdin")
	}
	return strings.TrimSpace(line), nil
}

func ensureBootstrapUser(name, email, password string) (*models.User, bool, error) {
	var users []models.User
	if err := facades.Orm().Query().Where("email = ?", email).Get(&users); err != nil {
		return nil, false, err
	}
	if len(users) > 0 {
		user := users[0]
		updates := map[string]any{"name": name, "status": "active"}
		if facades.Schema().HasColumn("users", "email_verified_at") {
			updates["email_verified_at"] = time.Now().UTC()
		}
		if _, err := facades.Orm().Query().Table("users").Where("id = ?", user.ID).Update(updates); err != nil {
			return nil, false, err
		}
		user.Name = name
		user.Status = "active"
		return &user, false, nil
	}

	user, err := userservices.NewUserService().Create(name, email, password, "zh-CN", "active")
	if err != nil {
		return nil, false, err
	}
	return user, true, nil
}

func ensureBootstrapRole(userID uint) error {
	var roles []models.Role
	if err := facades.Orm().Query().Where("name = ?", "super-admin").Get(&roles); err != nil {
		return err
	}
	var role models.Role
	if len(roles) == 0 {
		role = models.Role{Name: "super-admin", DisplayName: "Super Administrator"}
		if err := facades.Orm().Query().Create(&role); err != nil {
			return err
		}
	} else {
		role = roles[0]
	}

	assigned, err := facades.Orm().Query().Table("role_user").Where("user_id = ? AND role_id = ?", userID, role.ID).Exists()
	if err != nil {
		return err
	}
	if !assigned {
		if err := facades.Orm().Query().Table("role_user").Create(&map[string]any{"user_id": userID, "role_id": role.ID}); err != nil {
			return err
		}
	}

	for _, permission := range []models.Permission{
		{Name: "admin.users.view", DisplayName: "View users"},
		{Name: "admin.users.manage", DisplayName: "Manage users"},
		{Name: "admin.roles.manage", DisplayName: "Manage roles"},
		{Name: "admin.permissions.manage", DisplayName: "Manage permissions"},
		{Name: "admin.settings.manage", DisplayName: "Manage system settings"},
	} {
		var existing []models.Permission
		if err := facades.Orm().Query().Where("name = ?", permission.Name).Get(&existing); err != nil {
			return err
		}
		if len(existing) == 0 {
			if err := facades.Orm().Query().Create(&permission); err != nil {
				return err
			}
			existing = []models.Permission{permission}
		}
		if err := assignBootstrapPermission(existing[0].ID, role.ID); err != nil {
			return err
		}
	}

	var permissions []models.Permission
	if err := facades.Orm().Query().Model(&models.Permission{}).Get(&permissions); err != nil {
		return err
	}
	for _, permission := range permissions {
		if err := assignBootstrapPermission(permission.ID, role.ID); err != nil {
			return err
		}
	}
	return nil
}

func assignBootstrapPermission(permissionID, roleID uint) error {
	assigned, err := facades.Orm().Query().Table("permission_role").Where("permission_id = ? AND role_id = ?", permissionID, roleID).Exists()
	if err != nil {
		return err
	}
	if assigned {
		return nil
	}
	return facades.Orm().Query().Table("permission_role").Create(&map[string]any{"permission_id": permissionID, "role_id": roleID})
}
