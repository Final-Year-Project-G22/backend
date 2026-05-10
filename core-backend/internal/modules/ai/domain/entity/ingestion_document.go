package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type IngestionDocumentStatus string

const (
	IngestionDocumentStatusQueued     IngestionDocumentStatus = "queued"
	IngestionDocumentStatusValidating IngestionDocumentStatus = "validating"
	IngestionDocumentStatusFailed     IngestionDocumentStatus = "failed"
	IngestionDocumentStatusCompleted  IngestionDocumentStatus = "completed"
)

type IngestionDocument struct {
	model.BaseModel  `gorm:"embedded"`
	AccountID        uuid.UUID               `gorm:"type:uuid;not null;index:idx_ingestion_documents_account;uniqueIndex:idx_ingestion_docs_idempotency_per_account,priority:1"`
	UserID           uuid.UUID               `gorm:"type:uuid;not null;index:idx_ingestion_documents_user"`
	StorageKey       string                  `gorm:"type:text;not null;index:idx_ingestion_documents_storage_key"`
	ContentType      string                  `gorm:"type:varchar(255);not null"`
	SizeBytes        int64                   `gorm:"not null;default:0"`
	ChecksumSHA256   string                  `gorm:"type:varchar(128);not null"`
	IdempotencyKey   string                  `gorm:"type:varchar(255);not null;uniqueIndex:idx_ingestion_docs_idempotency_per_account,priority:2"`
	BatchID          *uuid.UUID              `gorm:"type:uuid;index:idx_ingestion_documents_batch"`
	SourceFilename   *string                 `gorm:"type:text"`
	DeclaredLanguage *string                 `gorm:"type:varchar(16)"`
	SchemaVersion    string                  `gorm:"type:varchar(32);not null;default:'1.0.0'"`
	Status           IngestionDocumentStatus `gorm:"type:varchar(32);not null;default:'queued';index:idx_ingestion_documents_status"`
	LastError        *string                 `gorm:"type:text"`
	EventID          uuid.UUID               `gorm:"type:uuid;not null;uniqueIndex:idx_ingestion_documents_event_id"`
	SectorIDs        []uuid.UUID             `gorm:"type:uuid[];index:idx_ingestion_documents_sector_ids,using:gin"`
	TagIDs           []uuid.UUID             `gorm:"type:uuid[];index:idx_ingestion_documents_tag_ids,using:gin"`
	Region           *string                 `gorm:"type:varchar(50)"`
	Stage            *string                 `gorm:"type:varchar(50)"`
}

func (IngestionDocument) TableName() string {
	return "ingestion_documents"
}
