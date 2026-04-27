package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationQueue struct {
	model.BaseModel  `gorm:"embedded"`
	NotificationType NotificationType     `gorm:"type:varchar(64);not null"`
	AccountID        uuid.UUID            `gorm:"type:uuid;not null;index:idx_notif_queue_account"`
	Priority         NotificationPriority `gorm:"type:smallint;not null;default:1"`
	TemplateID       *uuid.UUID           `gorm:"type:uuid"`
	Channel          Channel              `gorm:"type:varchar(20);not null"`
	Payload          datatypes.JSONMap    `gorm:"type:jsonb;not null"`
	ScheduledFor     time.Time            `gorm:"type:timestamptz;not null;index:idx_notif_queue_scheduled"`
	MaxRetries       int                  `gorm:"not null;default:3"`
	RetryCount       int                  `gorm:"not null;default:0"`
	Status           NotificationStatus   `gorm:"type:varchar(20);not null;default:'pending';index:idx_notif_queue_status"`
	ErrorMessage     *string              `gorm:"type:text"`
}

func (NotificationQueue) TableName() string {
	return "notification_queue"
}
