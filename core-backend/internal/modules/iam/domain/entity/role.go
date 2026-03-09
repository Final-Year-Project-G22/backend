package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

type RoleType string

const (
	RoleTypeSystem RoleType = "system"
	RoleTypeCustom RoleType = "custom"
)

type Role struct {
	model.BaseModel `gorm:"embedded"`

	Code            string           `gorm:"type:varchar(100);not null;uniqueIndex:idx_roles_code"`
	Name            string           `gorm:"type:varchar(100);not null"`
	Description     *string          `gorm:"type:text"`
	Type            RoleType         `gorm:"type:varchar(32);not null;default:'system'"`
	IsSystem        bool             `gorm:"not null;default:true"`
	IsMutable       bool             `gorm:"not null;default:false"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoleAssignments []RoleAssignment `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Role) TableName() string {
	return "roles"
}
