package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GuideStepTranslation struct {
	ID              uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GuideStepID     uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_step_trans,priority:1"`
	Language        string            `gorm:"type:varchar(10);not null;uniqueIndex:idx_step_trans,priority:2"`
	Title           string            `gorm:"type:varchar(200);not null"`
	Description     *string           `gorm:"type:text"`
	DetailedContent datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt       *time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *GuideStepTranslation) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (GuideStepTranslation) TableName() string {
	return "guide_step_translations"
}
