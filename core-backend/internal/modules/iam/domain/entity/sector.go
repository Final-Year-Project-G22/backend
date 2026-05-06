package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Sector struct {
	model.BaseModel `gorm:"embedded"`

	Slug         string              `gorm:"type:varchar(100);uniqueIndex;not null"`
	ParentID     *uuid.UUID          `gorm:"type:uuid;index"`
	Parent       *Sector             `gorm:"foreignKey:ParentID"`
	Translations []SectorTranslation `gorm:"foreignKey:SectorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Sector) TableName() string {
	return "sectors"
}

type SectorTranslation struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SectorID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_sector_trans,priority:1"`
	Language    constants.Locale `gorm:"type:varchar(10);not null;uniqueIndex:idx_sector_trans,priority:2"`
	Name        string           `gorm:"type:varchar(100);not null"`
	Description *string          `gorm:"type:text"`
	CreatedAt   *time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   *time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *SectorTranslation) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (SectorTranslation) TableName() string {
	return "sector_translations"
}
