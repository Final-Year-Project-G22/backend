package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

// TODO: remove when category endpoints are removed.
type LibraryCategory struct {
	model.BaseModel  `gorm:"embedded"`
	Name             string                       `gorm:"type:varchar(200);not null"`
	Slug             string                       `gorm:"type:varchar(200);not null;uniqueIndex:idx_library_categories_slug_per_parent,priority:2"`
	Icon             *string                      `gorm:"type:varchar(100)"`
	SortOrder        int                          `gorm:"not null;default:0"`
	ParentCategoryID *uuid.UUID                   `gorm:"type:uuid;uniqueIndex:idx_library_categories_slug_per_parent,priority:1;index:idx_library_categories_parent"`
	ParentCategory   *LibraryCategory             `gorm:"foreignKey:ParentCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ChildCategories  []LibraryCategory            `gorm:"foreignKey:ParentCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Translations     []LibraryCategoryTranslation `gorm:"foreignKey:LibraryCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsActive         bool                         `gorm:"not null;default:true"`
}

func (LibraryCategory) TableName() string {
	return "library_categories"
}
