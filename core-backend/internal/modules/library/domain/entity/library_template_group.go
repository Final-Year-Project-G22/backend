package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type LibraryTemplateGroup struct {
	model.BaseModel `gorm:"embedded"`
	Name            string            `gorm:"type:varchar(200);not null"`
	Description     *string           `gorm:"type:text"`
	Slug            string            `gorm:"type:varchar(200);not null;uniqueIndex:idx_library_template_groups_slug_per_cat,priority:2"`
	CategoryID      uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_library_template_groups_slug_per_cat,priority:1;index:idx_library_template_groups_category"`
	Category        LibraryCategory   `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Format          TemplateFormat    `gorm:"type:varchar(20);not null"`
	TierAccess      TierAccess        `gorm:"type:varchar(10);not null;default:'basic'"`
	RequiresAuth    bool              `gorm:"not null;default:true"`
	IsActive        bool              `gorm:"not null;default:true"`
	SortOrder       int               `gorm:"not null;default:0"`
	DefaultLanguage string            `gorm:"type:varchar(10);not null;default:'en'"`
	ThumbnailURL    *string           `gorm:"type:varchar(512)"`
	DownloadCount   int               `gorm:"not null;default:0"`
	CreatedBy       uuid.UUID         `gorm:"type:uuid;not null;index"`
	Templates       []LibraryTemplate `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (LibraryTemplateGroup) TableName() string {
	return "library_template_groups"
}
