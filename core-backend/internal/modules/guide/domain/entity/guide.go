package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

type Guide struct {
	model.BaseModel `gorm:"embedded"`
	SectorIDs       model.UUIDArray    `gorm:"type:uuid[];index:idx_guides_sector_ids,using:gin"`
	TagIDs          model.UUIDArray    `gorm:"type:uuid[];index:idx_guides_tag_ids,using:gin"`
	Slug            string             `gorm:"type:varchar(200);not null;uniqueIndex:idx_guides_slug"`
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
