package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserScheduledNotification struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_scheduled_account"`
	TemplateSlug    *string        `gorm:"type:varchar(64)"`
	Title           string         `gorm:"type:varchar(255);not null"`
	Body            string         `gorm:"type:text;not null"`
	Channels        pq.StringArray `gorm:"type:varchar(20)[];not null"`
	ScheduledFor    time.Time      `gorm:"type:timestamptz;not null;index:idx_user_scheduled_time"`
	Status          ScheduleStatus `gorm:"type:varchar(20);not null;default:'pending';index:idx_user_scheduled_status"`
	RescheduledFrom *time.Time     `gorm:"type:timestamptz"`
	SentAt          *time.Time     `gorm:"type:timestamptz"`
	CancelledAt     *time.Time     `gorm:"type:timestamptz"`
}

func (UserScheduledNotification) TableName() string {
	return "user_scheduled_notifications"
}
