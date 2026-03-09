package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

type Permission struct {
	model.BaseModel `gorm:"embedded"`

	Code            string           `gorm:"type:varchar(150);not null;uniqueIndex:idx_permissions_code"`
	Name            string           `gorm:"type:varchar(150);not null"`
	Description     *string          `gorm:"type:text"`
	Module          string           `gorm:"type:varchar(100);not null;index"`
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Permission) TableName() string {
	return "permissions"
}
