package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LibraryCategoryTranslation struct {
	ID                uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	LibraryCategoryID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_library_cat_trans,priority:1"`
	Language          string     `gorm:"type:varchar(10);not null;uniqueIndex:idx_library_cat_trans,priority:2"`
	Name              string     `gorm:"type:varchar(200);not null"`
	Description       *string    `gorm:"type:text"`
	CreatedAt         *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt         *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *LibraryCategoryTranslation) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (LibraryCategoryTranslation) TableName() string {
	return "library_category_translations"
}
