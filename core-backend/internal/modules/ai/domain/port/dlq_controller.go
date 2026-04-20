package port

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DLQController interface {
	ListDeadEvents(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*DLQEvent, error)
	GetDeadEvent(ctx context.Context, eventID uuid.UUID) (*DLQEvent, error)
	ReDriveEvent(ctx context.Context, eventID uuid.UUID, operatorID uuid.UUID) error
	ReDriveBatch(ctx context.Context, eventIDs []uuid.UUID, operatorID uuid.UUID) (int, error)
	GetReDriveHistory(ctx context.Context, eventID uuid.UUID) ([]DLQReDriveAudit, error)
}

type DLQEvent struct {
	EventID      uuid.UUID
	EventType    string
	Payload      datatypes.JSONMap
	Status       entity.OutboxStatus
	ErrorMessage *string
	CreatedAt    time.Time
	ReplayCount  int32
}

type DLQReDriveAudit struct {
	ID         uuid.UUID `json:"id"`
	EventID    uuid.UUID `json:"event_id"`
	OperatorID uuid.UUID `json:"operator_id"`
	ReDrivenAt time.Time `json:"re_driven_at"`
	Success    bool      `json:"success"`
	Error      *string   `json:"error,omitempty"`
}

type IngestionStatusEvent struct {
	DocumentID      uuid.UUID
	Status          string
	ErrorMessage    *string
	ProcessedAt     *time.Time
	TotalChunks     int
	ProcessedChunks int
}
