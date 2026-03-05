package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type AccountRole struct {
	model.BaseModel `gorm:"embedded"`

	AccountId uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account   Account   `gorm:"foreignKey:AccountId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	RoleId    uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_created_by_account_role_id"`
}

func (AccountRole) TableName() string {
	return "account_roles"
}
