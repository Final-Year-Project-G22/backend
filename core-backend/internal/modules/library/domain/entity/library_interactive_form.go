package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type LibraryInteractiveForm struct {
	model.BaseModel `gorm:"embedded"`
	TemplateID      uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex:idx_library_interactive_forms_template"`
	Template        LibraryTemplate        `gorm:"foreignKey:TemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name            string                 `gorm:"type:varchar(100);not null"`
	Description     *string                `gorm:"type:text"`
	FormLayout      map[string]interface{} `gorm:"type:jsonb;not null"`
	Version         int                    `gorm:"not null;default:1"`
	IsActive        bool                   `gorm:"not null;default:true"`
}

func (LibraryInteractiveForm) TableName() string {
	return "library_interactive_forms"
}
