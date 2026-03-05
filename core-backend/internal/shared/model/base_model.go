package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt *time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt *gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate ensures a UUID is assigned
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
