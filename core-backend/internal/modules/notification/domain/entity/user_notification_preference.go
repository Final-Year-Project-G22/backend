package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserNotificationPreference struct {
	model.BaseModel  `gorm:"embedded"`
	AccountID        uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_user_notif_prefs_account_type_channel,priority:1;index:idx_user_notif_prefs_account"`
	NotificationType NotificationType `gorm:"type:varchar(64);not null;uniqueIndex:idx_user_notif_prefs_account_type_channel,priority:2"`
	Channel          Channel          `gorm:"type:varchar(20);not null;uniqueIndex:idx_user_notif_prefs_account_type_channel,priority:3"`
	IsEnabled        bool             `gorm:"not null"`
	QuietHoursStart  *time.Time       `gorm:"type:time"`
	QuietHoursEnd    *time.Time       `gorm:"type:time"`
}

func (UserNotificationPreference) TableName() string {
	return "user_notification_preferences"
}
