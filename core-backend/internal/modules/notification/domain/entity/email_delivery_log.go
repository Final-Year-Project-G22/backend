package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type EmailDeliveryLog struct {
	model.BaseModel       `gorm:"embedded"`
	NotificationHistoryID uuid.UUID  `gorm:"type:uuid;not null"`
	Provider              string     `gorm:"type:varchar(50);not null"`
	ProviderMessageID     *string    `gorm:"type:varchar(255);index:idx_email_delivery_provider_msg"`
	RecipientEmail        string     `gorm:"type:varchar(255);not null"`
	Subject               string     `gorm:"type:varchar(500);not null"`
	SentAt                time.Time  `gorm:"type:timestamptz;not null"`
	DeliveredAt           *time.Time `gorm:"type:timestamptz"`
	OpenedAt              *time.Time `gorm:"type:timestamptz"`
	ClickedAt             *time.Time `gorm:"type:timestamptz"`
	BounceReason          *string    `gorm:"type:text"`
	Complaint             bool       `gorm:"not null;default:false"`
}

func (EmailDeliveryLog) TableName() string {
	return "email_delivery_logs"
}
