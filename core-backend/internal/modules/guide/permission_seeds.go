package guide

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func GuideSeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		{
			Code:        GuideRead,
			Name:        GuideRead,
			Description: permissions.StringPtr("Read access to guide resources"),
			Module:      "guide",
		},
		{
			Code:        GuideWrite,
			Name:        GuideWrite,
			Description: permissions.StringPtr("Create guide resources"),
			Module:      "guide",
		},
		{
			Code:        GuideUpdate,
			Name:        GuideUpdate,
			Description: permissions.StringPtr("Update guide resources"),
			Module:      "guide",
		},
		{
			Code:        GuideDelete,
			Name:        GuideDelete,
			Description: permissions.StringPtr("Delete guide resources"),
			Module:      "guide",
		},
	}
}

func GuideSeedRoles() []permissions.SeedRole { return nil }
