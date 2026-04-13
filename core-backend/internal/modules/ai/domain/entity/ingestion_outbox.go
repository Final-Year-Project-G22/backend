package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"
)

type IngestionOutbox struct {
	model.BaseModel `gorm:"embedded"`
	EventID         uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_ingestion_outbox_event_id"`
	EventType       string            `gorm:"type:varchar(255);not null;index:idx_ingestion_outbox_event_type;uniqueIndex:idx_ingestion_outbox_dedupe,priority:1"`
	SchemaVersion   string            `gorm:"type:varchar(32);not null;default:'1.0.0'"`
	Producer        string            `gorm:"type:varchar(64);not null"`
	KeyID           string            `gorm:"type:varchar(128);not null"`
	IdempotencyKey  string            `gorm:"type:varchar(255);not null;uniqueIndex:idx_ingestion_outbox_dedupe,priority:2"`
	AggregateID     uuid.UUID         `gorm:"type:uuid;not null;index:idx_ingestion_outbox_aggregate"`
	AccountID       uuid.UUID         `gorm:"type:uuid;not null;index:idx_ingestion_outbox_account"`
	UserID          uuid.UUID         `gorm:"type:uuid;not null;index:idx_ingestion_outbox_user"`
	BatchID         *uuid.UUID        `gorm:"type:uuid;index:idx_ingestion_outbox_batch"`
	ReplayCount     int32             `gorm:"not null;default:0"`
	Payload         datatypes.JSONMap `gorm:"type:jsonb;not null"`
	Signature       []byte            `gorm:"type:bytea"`
	Status          OutboxStatus      `gorm:"type:varchar(32);not null;default:'pending';index:idx_ingestion_outbox_status_next_attempt,priority:1"`
	AttemptCount    int               `gorm:"not null;default:0"`
	NextAttemptAt   *time.Time        `gorm:"type:timestamptz;index:idx_ingestion_outbox_status_next_attempt,priority:2"`
	PublishedAt     *time.Time        `gorm:"type:timestamptz"`
	LastError       *string           `gorm:"type:text"`
}

func (IngestionOutbox) TableName() string {
	return "ingestion_outbox"
}
