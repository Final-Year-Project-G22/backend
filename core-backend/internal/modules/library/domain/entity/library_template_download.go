package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type LibraryTemplateDownload struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TemplateID      uuid.UUID `gorm:"type:uuid;not null;index"`
	GroupID         uuid.UUID `gorm:"type:uuid;not null;index"`
}

func (LibraryTemplateDownload) TableName() string {
	return "library_template_downloads"
}
