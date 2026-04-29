package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type LibraryTemplate struct {
	model.BaseModel `gorm:"embedded"`
	GroupID         uuid.UUID               `gorm:"type:uuid;not null;uniqueIndex:idx_library_templates_group_lang,priority:1;index:idx_library_templates_group"`
	Group           LibraryTemplateGroup    `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Language        string                  `gorm:"type:varchar(10);not null;uniqueIndex:idx_library_templates_group_lang,priority:2"`
	Title           string                  `gorm:"type:varchar(200);not null"`
	Description     *string                 `gorm:"type:text"`
	FileKey         string                  `gorm:"type:varchar(512);not null"`
	FileURL         *string                 `gorm:"type:varchar(512)"`
	FileSize        int64                   `gorm:"not null"`
	ContentType     string                  `gorm:"type:varchar(100);not null"`
	Version         int                     `gorm:"not null;default:1"`
	IsActive        bool                    `gorm:"not null;default:true"`
	InteractiveForm *LibraryInteractiveForm `gorm:"foreignKey:TemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (LibraryTemplate) TableName() string {
	return "library_templates"
}
