package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type TemplatePreference struct {
	model.BaseModel `gorm:"embedded"`

	AccountID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account         Account   `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DefaultTemplate *string   `gorm:"type:varchar(100)"`
	EditorMode      *string   `gorm:"type:varchar(50)"`
}

func (TemplatePreference) TableName() string {
	return "template_preferences"
}
