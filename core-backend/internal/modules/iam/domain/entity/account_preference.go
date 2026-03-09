package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type AccountPreference struct {
	model.BaseModel `gorm:"embedded"`

	AccountID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account   Account   `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Language  string    `gorm:"type:varchar(10);not null;default:'en'"`
	Timezone  string    `gorm:"type:varchar(64);not null;default:'UTC'"`
}

func (AccountPreference) TableName() string {
	return "account_preferences"
}
