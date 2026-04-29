package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserNotificationInbox struct {
	model.BaseModel       `gorm:"embedded"`
	AccountID             uuid.UUID            `gorm:"type:uuid;not null;index:idx_notif_inbox_account"`
	NotificationHistoryID uuid.UUID            `gorm:"type:uuid;not null"`
	Category              NotificationCategory `gorm:"type:varchar(32);not null;index:idx_notif_inbox_category"`
	ActionUrl             *string              `gorm:"type:varchar(512)"`
	IsRead                bool                 `gorm:"not null;default:false"`
	IsArchived            bool                 `gorm:"not null;default:false"`
	ExpiresAt             *time.Time           `gorm:"type:timestamptz;index:idx_notif_inbox_expires"`
	NotificationHistory   NotificationHistory  `gorm:"foreignKey:NotificationHistoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (UserNotificationInbox) TableName() string {
	return "user_notification_inboxes"
}
