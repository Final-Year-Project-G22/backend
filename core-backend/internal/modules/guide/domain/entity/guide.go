package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type Guide struct {
	model.BaseModel `gorm:"embedded"`
	CategoryID      uuid.UUID          `gorm:"type:uuid;not null;uniqueIndex:idx_guides_slug_per_category,priority:1;index:idx_guides_category"`
	Category        GuideCategory      `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Slug            string             `gorm:"type:varchar(200);not null;uniqueIndex:idx_guides_slug_per_category,priority:2"`
	Icon            *string            `gorm:"type:varchar(100)"`
	SortOrder       int                `gorm:"not null;default:0"`
	Conditions      []GuideCondition   `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Steps           []GuideStep        `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Journeys        []UserGuideJourney `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Translations    []GuideTranslation `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Guide) TableName() string {
	return "guides"
}
