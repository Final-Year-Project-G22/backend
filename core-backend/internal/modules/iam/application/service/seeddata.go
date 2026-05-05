package service

import (
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
)

type seedPermission struct {
	Code        string
	Name        string
	Description *string
	Module      string
}

type seedRole struct {
	Code        string
	Name        string
	Description *string
}

func seedPermissions() []seedPermission {
	return []seedPermission{
		// Generic IAM permissions
		{
			Code:        permissions.IAMRead,
			Name:        permissions.IAMRead,
			Description: stringPtr("Read access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        permissions.IAMWrite,
			Name:        permissions.IAMWrite,
			Description: stringPtr("Write access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        permissions.IAMUpdate,
			Name:        permissions.IAMUpdate,
			Description: stringPtr("Update access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        permissions.IAMDelete,
			Name:        permissions.IAMDelete,
			Description: stringPtr("Delete access to IAM resources"),
			Module:      "iam",
		},
		// Fine-grained admin management permissions
		{
			Code:        permissions.AdminList,
			Name:        permissions.AdminList,
			Description: stringPtr("List all admin accounts"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminRead,
			Name:        permissions.AdminRead,
			Description: stringPtr("View detailed information about an admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminCreate,
			Name:        permissions.AdminCreate,
			Description: stringPtr("Register a new admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminRolesUpdate,
			Name:        permissions.AdminRolesUpdate,
			Description: stringPtr("Update roles assigned to an admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminResetPassword,
			Name:        permissions.AdminResetPassword,
			Description: stringPtr("Trigger a password reset for an admin account"),
			Module:      "iam",
		},
		{
			Code:        permissions.AdminStatusUpdate,
			Name:        permissions.AdminStatusUpdate,
			Description: stringPtr("Lock, suspend, or activate an admin account"),
			Module:      "iam",
		},
		// Role management permissions
		{
			Code:        permissions.RoleRead,
			Name:        permissions.RoleRead,
			Description: stringPtr("View role details and permissions"),
			Module:      "iam",
		},
		{
			Code:        permissions.RoleCreate,
			Name:        permissions.RoleCreate,
			Description: stringPtr("Create a new custom role"),
			Module:      "iam",
		},
		{
			Code:        permissions.RoleUpdate,
			Name:        permissions.RoleUpdate,
			Description: stringPtr("Update an existing custom role"),
			Module:      "iam",
		},
		{
			Code:        permissions.RoleDelete,
			Name:        permissions.RoleDelete,
			Description: stringPtr("Delete a custom role"),
			Module:      "iam",
		},
		// Permission catalog
		{
			Code:        permissions.PermissionRead,
			Name:        permissions.PermissionRead,
			Description: stringPtr("List and view all available permissions"),
			Module:      "iam",
		},
	}
}

func seedRoles() []seedRole {
	return []seedRole{
		{
			Code:        "super_admin",
			Name:        "Super Admin",
			Description: stringPtr("Full access to all IAM capabilities"),
		},
		{
			Code:        "iam_admin",
			Name:        "IAM Admin",
			Description: stringPtr("Administrative access to IAM"),
		},
	}
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
