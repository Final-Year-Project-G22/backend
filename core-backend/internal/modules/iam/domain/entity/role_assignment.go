package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type RoleAssignment struct {
	model.BaseModel `gorm:"embedded"`

	AccountID         uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_role_assignments_account_role,priority:1"`
	Account           Account    `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoleID            uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_role_assignments_account_role,priority:2"`
	Role              Role       `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AssignedBy        uuid.UUID  `gorm:"type:uuid;not null;index"`
	AssignedByAccount Account    `gorm:"foreignKey:AssignedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ExpiresAt         *time.Time `gorm:"type:timestamptz"`
	RevokedAt         *time.Time `gorm:"type:timestamptz"`
	RevokeReason      *string    `gorm:"type:text"`
}

func (RoleAssignment) TableName() string {
	return "role_assignments"
}
