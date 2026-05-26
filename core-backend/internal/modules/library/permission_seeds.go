package library

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func LibrarySeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		{
			Code:        LibraryRead,
			Name:        LibraryRead,
			Description: permissions.StringPtr("Read access to library resources"),
			Module:      "library",
		},
		{
			Code:        LibraryWrite,
			Name:        LibraryWrite,
			Description: permissions.StringPtr("Create library resources"),
			Module:      "library",
		},
		{
			Code:        LibraryUpdate,
			Name:        LibraryUpdate,
			Description: permissions.StringPtr("Update library resources"),
			Module:      "library",
		},
		{
			Code:        LibraryDelete,
			Name:        LibraryDelete,
			Description: permissions.StringPtr("Delete library resources"),
			Module:      "library",
		},
	}
}

func LibrarySeedRoles() []permissions.SeedRole {
	return []permissions.SeedRole{
		{
			Code:        "library_admin",
			Name:        "Library Admin",
			Description: permissions.StringPtr("Administrative access to library resources"),
			PermissionCodes: []string{
				LibraryRead,
				LibraryWrite,
				LibraryUpdate,
				LibraryDelete,
			},
		},
	}
}
