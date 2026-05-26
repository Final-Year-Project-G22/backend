package service

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func IAMSeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		// Generic IAM permissions
		{
			Code:        permissions.IAMRead,
			Name:        permissions.IAMRead,
			Description: permissions.StringPtr("Read access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        permissions.IAMWrite,
			Name:        permissions.IAMWrite,
			Description: permissions.StringPtr("Write access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        permissions.IAMUpdate,
			Name:        permissions.IAMUpdate,
			Description: permissions.StringPtr("Update access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        permissions.IAMDelete,
			Name:        permissions.IAMDelete,
			Description: permissions.StringPtr("Delete access to IAM resources"),
			Module:      "iam",
		},
		// Fine-grained admin management permissions
		{
			Code:        permissions.AdminList,
			Name:        permissions.AdminList,
			Description: permissions.StringPtr("List all admin accounts"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminRead,
			Name:        permissions.AdminRead,
			Description: permissions.StringPtr("View detailed information about an admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminCreate,
			Name:        permissions.AdminCreate,
			Description: permissions.StringPtr("Register a new admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminRolesUpdate,
			Name:        permissions.AdminRolesUpdate,
			Description: permissions.StringPtr("Update roles assigned to an admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminResetPassword,
			Name:        permissions.AdminResetPassword,
			Description: permissions.StringPtr("Trigger a password reset for an admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminStatusUpdate,
			Name:        permissions.AdminStatusUpdate,
			Description: permissions.StringPtr("Lock, suspend, or activate an admin account"),
			Module:      "iam",
		},
		// Role management permissions
		{
			Code:        permissions.RoleRead,
			Name:        permissions.RoleRead,
			Description: permissions.StringPtr("View role details and permissions"),
			Module:      "iam",
		},
		{
			Code:        permissions.RoleCreate,
			Name:        permissions.RoleCreate,
			Description: permissions.StringPtr("Create a new custom role"),
			Module:      "iam",
		},
		{
			Code:        permissions.RoleUpdate,
			Name:        permissions.RoleUpdate,
			Description: permissions.StringPtr("Update an existing custom role"),
			Module:      "iam",
		},
		{
			Code:        permissions.RoleDelete,
			Name:        permissions.RoleDelete,
			Description: permissions.StringPtr("Delete a custom role"),
			Module:      "iam",
		},
		// Permission catalog
		{
			Code:        permissions.PermissionRead,
			Name:        permissions.PermissionRead,
			Description: permissions.StringPtr("List and view all available permissions"),
			Module:      "iam",
		},
	}
}

func IAMSeedRoles() []permissions.SeedRole {
	return []permissions.SeedRole{
		{
			Code:        "super_admin",
			Name:        "Super Admin",
			Description: permissions.StringPtr("Full access to all system capabilities"),
		},
		{
			Code:        "iam_admin",
			Name:        "IAM Admin",
			Description: permissions.StringPtr("Administrative access to IAM"),
			PermissionCodes: []string{
				"iam.read",
				"iam.write",
				"iam.update",
				"iam.delete",
				"iam.admin.list",
				"iam.admin.read",
				"iam.admin.create",
				"iam.admin.roles.update",
				"iam.role.read",
				"iam.role.create",
				"iam.role.update",
				"iam.permission.read",
			},
		},
	}
}
