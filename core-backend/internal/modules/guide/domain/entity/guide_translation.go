package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GuideTranslation struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GuideID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_guide_trans,priority:1"`
	Language    string     `gorm:"type:varchar(10);not null;uniqueIndex:idx_guide_trans,priority:2"`
	Name        string     `gorm:"type:varchar(200);not null"`
	Description *string    `gorm:"type:text"`
	CreatedAt   *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *GuideTranslation) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (GuideTranslation) TableName() string {
	return "guide_translations"
}
