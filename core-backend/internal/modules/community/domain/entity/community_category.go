package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type CommunityCategory struct {
	model.BaseModel  `gorm:"embedded"`
	Name             string              `gorm:"type:varchar(200);not null"`
	Slug             string              `gorm:"type:varchar(200);not null;uniqueIndex:idx_community_categories_slug_per_parent,priority:2"`
	Description      *string             `gorm:"type:text"`
	ParentCategoryID *uuid.UUID          `gorm:"type:uuid;uniqueIndex:idx_community_categories_slug_per_parent,priority:1;index:idx_community_categories_parent"`
	ParentCategory   *CommunityCategory  `gorm:"foreignKey:ParentCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ChildCategories  []CommunityCategory `gorm:"foreignKey:ParentCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Threads          []DiscussionThread  `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsActive         bool                `gorm:"not null;default:true"`
}

func (CommunityCategory) TableName() string {
	return "community_categories"
}
