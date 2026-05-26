package ai

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func AISeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		{
			Code:        AIAdminStream,
			Name:        AIAdminStream,
			Description: permissions.StringPtr("Access the AI admin debug streaming endpoint with full ReAct loop visibility"),
			Module:      "ai",
		},
	}
}

func AISeedRoles() []permissions.SeedRole { return nil }
