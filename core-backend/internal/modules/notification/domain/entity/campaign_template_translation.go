package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CampaignTemplateTranslation struct {
	ID                 uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CampaignTemplateID uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_ctrans_template_lang,priority:1"`
	Language           string            `gorm:"type:varchar(10);not null;uniqueIndex:idx_ctrans_template_lang,priority:2"`
	Content            datatypes.JSONMap `gorm:"type:jsonb;not null"`
	CreatedAt          *time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt          *time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *CampaignTemplateTranslation) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (CampaignTemplateTranslation) TableName() string {
	return "campaign_template_translations"
}
