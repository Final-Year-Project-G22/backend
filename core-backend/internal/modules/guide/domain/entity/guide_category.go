package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type GuideCategory struct {
	model.BaseModel  `gorm:"embedded"`
	Slug             string                     `gorm:"type:varchar(200);not null;uniqueIndex:idx_guide_categories_slug_per_parent,priority:2"`
	Icon             *string                    `gorm:"type:varchar(100)"`
	SortOrder        int                        `gorm:"not null;default:0"`
	ParentCategoryID *uuid.UUID                 `gorm:"type:uuid;uniqueIndex:idx_guide_categories_slug_per_parent,priority:1;index:idx_guide_categories_parent"`
	ParentCategory   *GuideCategory             `gorm:"foreignKey:ParentCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ChildCategories  []GuideCategory            `gorm:"foreignKey:ParentCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Guides           []Guide                    `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Conditions       []GuideCategoryCondition   `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Translations     []GuideCategoryTranslation `gorm:"foreignKey:GuideCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (GuideCategory) TableName() string {
	return "guide_categories"
}
