package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type Session struct {
	model.BaseModel `gorm:"embedded"`

	AccountID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	Account          Account    `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RefreshTokenHash string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_sessions_refresh_token_hash"`
	UserAgent        *string    `gorm:"type:text"`
	IPAddress        *string    `gorm:"type:varchar(64)"`
	LastActiveAt     time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt        time.Time  `gorm:"type:timestamptz;not null;index"`
	RevokedAt        *time.Time `gorm:"type:timestamptz"`
}

func (Session) TableName() string {
	return "sessions"
}
