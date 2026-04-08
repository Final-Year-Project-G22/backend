package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserThreadSettings struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_user_thread_settings_account_thread,priority:1;index:idx_user_thread_settings_account"`
	ThreadID        uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_user_thread_settings_account_thread,priority:2;index:idx_user_thread_settings_thread"`
	Thread          DiscussionThread `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsFollowing     bool             `gorm:"not null;default:true"`
	IsMuted         bool             `gorm:"not null;default:false"`
	LastReadAt      *time.Time       `gorm:"type:timestamptz"`
}

func (UserThreadSettings) TableName() string {
	return "user_thread_settings"
}
