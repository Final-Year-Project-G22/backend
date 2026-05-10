package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationOutboxStatus string

const (
	NotificationOutboxStatusPending    NotificationOutboxStatus = "pending"
	NotificationOutboxStatusPublished  NotificationOutboxStatus = "published"
	NotificationOutboxStatusDeadLetter NotificationOutboxStatus = "dead_letter"
)

type NotificationOutbox struct {
	model.BaseModel `gorm:"embedded"`
	EventType       string                   `gorm:"type:varchar(255);not null"`
	SchemaVersion   string                   `gorm:"type:varchar(32);not null;default:'1.0.0'"`
	SourceModule    string                   `gorm:"type:varchar(64);not null;index:idx_notification_outbox_source"`
	AccountID       uuid.UUID                `gorm:"type:uuid;not null;index:idx_notification_outbox_account"`
	IdempotencyKey  string                   `gorm:"type:varchar(255);not null;uniqueIndex:idx_notification_outbox_idempotency"`
	Payload         datatypes.JSONMap        `gorm:"type:jsonb;not null"`
	Status          NotificationOutboxStatus `gorm:"type:varchar(32);not null;default:'pending';index:idx_notification_outbox_status_next,priority:1"`
	AttemptCount    int                      `gorm:"not null;default:0"`
	NextAttemptAt   *time.Time               `gorm:"type:timestamptz;index:idx_notification_outbox_status_next,priority:2"`
	PublishedAt     *time.Time               `gorm:"type:timestamptz"`
	LastError       *string                  `gorm:"type:text"`
}

func (NotificationOutbox) TableName() string {
	return "notification_outbox"
}
