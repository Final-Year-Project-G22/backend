package community

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func CommunitySeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		{
			Code:        CommunityRead,
			Name:        CommunityRead,
			Description: permissions.StringPtr("Read access to community resources"),
			Module:      "community",
		},
		{
			Code:        CommunityWrite,
			Name:        CommunityWrite,
			Description: permissions.StringPtr("Create community resources"),
			Module:      "community",
		},
		{
			Code:        CommunityUpdate,
			Name:        CommunityUpdate,
			Description: permissions.StringPtr("Update community resources"),
			Module:      "community",
		},
		{
			Code:        CommunityDelete,
			Name:        CommunityDelete,
			Description: permissions.StringPtr("Delete community resources"),
			Module:      "community",
		},
	}
}

func CommunitySeedRoles() []permissions.SeedRole {
	return []permissions.SeedRole{
		{
			Code:        "community_admin",
			Name:        "Community Admin",
			Description: permissions.StringPtr("Administrative access to community resources"),
			PermissionCodes: []string{
				CommunityRead,
				CommunityWrite,
				CommunityUpdate,
				CommunityDelete,
			},
		},
	}
}
