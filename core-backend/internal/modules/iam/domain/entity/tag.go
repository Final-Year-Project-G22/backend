package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	model.BaseModel `gorm:"embedded"`

	Slug          string           `gorm:"type:varchar(100);uniqueIndex;not null"`
	Group         TagGroup         `gorm:"type:varchar(50);not null;index"`
	IsMultiSelect bool             `gorm:"type:boolean;not null;default:true"`
	Translations  []TagTranslation `gorm:"foreignKey:TagID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Tag) TableName() string {
	return "tags"
}

type TagTranslation struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TagID       uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_tag_trans,priority:1"`
	Language    constants.Locale `gorm:"type:varchar(10);not null;uniqueIndex:idx_tag_trans,priority:2"`
	Name        string           `gorm:"type:varchar(100);not null"`
	Description *string          `gorm:"type:text"`
	CreatedAt   *time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   *time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *TagTranslation) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (TagTranslation) TableName() string {
	return "tag_translations"
}
