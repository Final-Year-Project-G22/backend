package service

import "strings"

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
			Code:        "read",
			Name:        "Read access",
			Description: stringPtr("Read access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        "write",
			Name:        "Write access",
			Description: stringPtr("Write access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        "update",
			Name:        "Update access",
			Description: stringPtr("Update access to IAM resources"),
			Module:      "iam",
		},
		{
			Code:        "delete",
			Name:        "Delete access",
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
