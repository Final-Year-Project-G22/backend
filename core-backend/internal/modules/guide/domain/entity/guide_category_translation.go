package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GuideCategoryTranslation struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GuideCategoryID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_cat_trans,priority:1"`
	Language        string     `gorm:"type:varchar(10);not null;uniqueIndex:idx_cat_trans,priority:2"`
	Name            string     `gorm:"type:varchar(200);not null"`
	Description     *string    `gorm:"type:text"`
	CreatedAt       *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *GuideCategoryTranslation) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (GuideCategoryTranslation) TableName() string {
	return "guide_category_translations"
}
