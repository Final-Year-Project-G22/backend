package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type ThreadBlockedUser struct {
	model.BaseModel    `gorm:"embedded"`
	ThreadID           uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_thread_blocked_users_thread_account,priority:1;index:idx_thread_blocked_users_thread"`
	Thread             DiscussionThread `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BlockedAccountID   uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_thread_blocked_users_thread_account,priority:2;index:idx_thread_blocked_users_blocked"`
	BlockedByAccountID uuid.UUID        `gorm:"type:uuid;not null;index:idx_thread_blocked_users_blocked_by"`
	Reason             *string          `gorm:"type:text"`
}

func (ThreadBlockedUser) TableName() string {
	return "thread_blocked_users"
}
