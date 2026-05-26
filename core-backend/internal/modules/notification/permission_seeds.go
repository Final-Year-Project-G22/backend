package notification

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"

func NotificationSeedPermissions() []permissions.SeedPermission {
	return []permissions.SeedPermission{
		{
			Code:        NotificationRead,
			Name:        NotificationRead,
			Description: permissions.StringPtr("Read access to notification resources"),
			Module:      "notification",
		},
		{
			Code:        NotificationWrite,
			Name:        NotificationWrite,
			Description: permissions.StringPtr("Create notification resources"),
			Module:      "notification",
		},
		{
			Code:        NotificationUpdate,
			Name:        NotificationUpdate,
			Description: permissions.StringPtr("Update notification resources"),
			Module:      "notification",
		},
		{
			Code:        NotificationDelete,
			Name:        NotificationDelete,
			Description: permissions.StringPtr("Delete notification resources"),
			Module:      "notification",
		},
	}
}

func NotificationSeedRoles() []permissions.SeedRole { return nil }
