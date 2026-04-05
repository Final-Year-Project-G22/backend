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
		{
			Code:        "iam.read",
			Name:        permissions.ReadAccess,
			Description: stringPtr("Read access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        "iam.write",
			Name:        permissions.WriteAccess,
			Description: stringPtr("Write access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        "iam.update",
			Name:        permissions.UpdateAccess,
			Description: stringPtr("Update access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        "iam.delete",
			Name:        permissions.DeleteAccess,
			Description: stringPtr("Delete access to IAM resources"),
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
