package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"gorm.io/datatypes"
)

type CampaignTemplate struct {
	model.BaseModel `gorm:"embedded"`
	Name            string                        `gorm:"type:varchar(200);not null;uniqueIndex"`
	Description     *string                       `gorm:"type:text"`
	DefaultContent  datatypes.JSONMap             `gorm:"type:jsonb;not null"`
	Translations    []CampaignTemplateTranslation `gorm:"foreignKey:CampaignTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (CampaignTemplate) TableName() string {
	return "campaign_templates"
}
