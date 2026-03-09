package entity

import "github.com/google/uuid"

type RolePermission struct {
	RoleID       uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_role_permissions_role_permission,priority:1"`
	Role         Role       `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PermissionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_role_permissions_role_permission,priority:2"`
	Permission   Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
