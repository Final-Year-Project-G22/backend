package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type AIPreference struct {
	model.BaseModel `gorm:"embedded"`

	AccountID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account            Account   `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DefaultModel       *string   `gorm:"type:varchar(100)"`
	ResponseStyle      *string   `gorm:"type:varchar(100)"`
	Temperature        *float64
	AllowDataRetention bool `gorm:"not null;default:false"`
}

func (AIPreference) TableName() string {
	return "ai_preferences"
}
