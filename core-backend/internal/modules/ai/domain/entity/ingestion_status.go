package entity

import (
	"time"

	"github.com/google/uuid"
)

type IngestionStage string

const (
	IngestionStageQueued     IngestionStage = "queued"
	IngestionStageValidating IngestionStage = "validating"
	IngestionStageFetching   IngestionStage = "fetching"
	IngestionStageChunking   IngestionStage = "chunking"
	IngestionStageEmbedding  IngestionStage = "embedding"
	IngestionStageIndexing   IngestionStage = "indexing"
	IngestionStageCompleted  IngestionStage = "completed"
	IngestionStageFailed     IngestionStage = "failed"
	IngestionStageCancelled  IngestionStage = "cancelled"
)

func (s IngestionStage) IsTerminal() bool {
	return s == IngestionStageCompleted ||
		s == IngestionStageFailed ||
		s == IngestionStageCancelled
}

type IngestionStatusEvent struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID       string    `gorm:"type:varchar(100);not null;index"`
	DocumentID    uuid.UUID `gorm:"type:uuid;not null;index:idx_status_document_occurred"`
	AccountID     uuid.UUID `gorm:"type:uuid;not null;index:idx_status_account"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_status_user"`
	EventType     string    `gorm:"type:varchar(100);not null"`
	SchemaVersion string    `gorm:"type:varchar(32);not null"`
	OccurredAt    time.Time `gorm:"type:timestamp with time zone;not null;index:idx_status_document_occurred"`
	CreatedAt     time.Time `gorm:"type:timestamp with time zone;not null;autoCreateTime"`

	FromStage  *IngestionStage `gorm:"type:varchar(32)"`
	ToStage    IngestionStage  `gorm:"type:varchar(32);not null"`
	IsTerminal bool            `gorm:"not null;default:false"`
	RetryCount int             `gorm:"not null;default:0"`

	ErrorMessage         *string `gorm:"type:text"`
	ChunksProcessedCount *int    `gorm:"not null"`
	ChunksFailedCount    *int    `gorm:"not null"`

	EventSequence int64 `gorm:"not null;default:0;index"`
}

func (IngestionStatusEvent) TableName() string {
	return "ingestion_status_events"
}

type IngestionStatusProjection struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	DocumentID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_projection_document"`
	AccountID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_projection_account"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_projection_user"`
	EventID      string         `gorm:"type:varchar(100);not null"`
	CurrentStage IngestionStage `gorm:"type:varchar(32);not null"`
	IsTerminal   bool           `gorm:"not null;default:false"`
	StartedAt    time.Time      `gorm:"type:timestamp with time zone;not null"`
	UpdatedAt    time.Time      `gorm:"type:timestamp with time zone;not null;autoUpdateTime"`
	CompletedAt  *time.Time     `gorm:"type:timestamp with time zone"`
	LastError    *string        `gorm:"type:text"`

	ChunksProcessedCount int `gorm:"not null;default:0"`
	ChunksFailedCount    int `gorm:"not null;default:0"`

	LastEventSequence int64 `gorm:"not null;default:0"`
}

func (IngestionStatusProjection) TableName() string {
	return "ingestion_status_projections"
}
