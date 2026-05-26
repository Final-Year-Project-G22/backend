package ai

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func AISeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		{
			Code:        AIRead,
			Name:        AIRead,
			Description: permissions.StringPtr("View AI resources including knowledge base documents and status"),
			Module:      "ai",
		},
		{
			Code:        AIWrite,
			Name:        AIWrite,
			Description: permissions.StringPtr("Create, update, and delete AI resources including knowledge base documents"),
			Module:      "ai",
		},
		{
			Code:        AIAdminStream,
			Name:        AIAdminStream,
			Description: permissions.StringPtr("Access the AI admin debug streaming endpoint with full ReAct loop visibility"),
			Module:      "ai",
		},
	}
}

func AISeedRoles() []permissions.SeedRole { return nil }
