package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationHistory struct {
	model.BaseModel  `gorm:"embedded"`
	AccountID        uuid.UUID          `gorm:"type:uuid;not null;index:idx_notif_history_account"`
	NotificationType NotificationType   `gorm:"type:varchar(64);not null"`
	Channel          Channel            `gorm:"type:varchar(20);not null"`
	Title            string             `gorm:"type:varchar(500);not null"`
	Content          string             `gorm:"type:text;not null"`
	ActionUrl        *string            `gorm:"type:varchar(512)"`
	SentAt           time.Time          `gorm:"type:timestamptz;not null"`
	DeliveredAt      *time.Time         `gorm:"type:timestamptz"`
	ReadAt           *time.Time         `gorm:"type:timestamptz"`
	ClickedAt        *time.Time         `gorm:"type:timestamptz"`
	DeliveryStatus   DeliveryStatus     `gorm:"type:varchar(20);not null"`
	FailureReason    *string            `gorm:"type:text"`
	Metadata         *datatypes.JSONMap `gorm:"type:jsonb"`
}

func (NotificationHistory) TableName() string {
	return "notification_histories"
}
