package service

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/google/uuid"
)

type DLQController interface {
	ListDeadEvents(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*port.DLQEvent, error)
	GetDeadEvent(ctx context.Context, eventID uuid.UUID) (*port.DLQEvent, error)
	ReDriveEvent(ctx context.Context, eventID uuid.UUID, operatorID uuid.UUID) error
	ReDriveBatch(ctx context.Context, eventIDs []uuid.UUID, operatorID uuid.UUID) (int, error)
}

type dlqControllerAdapter struct {
	delegate port.DLQController
}

func NewDLQControllerAdapter(delegate port.DLQController) DLQController {
	return &dlqControllerAdapter{delegate: delegate}
}

func (a *dlqControllerAdapter) ListDeadEvents(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*port.DLQEvent, error) {
	events, err := a.delegate.ListDeadEvents(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*port.DLQEvent, 0, len(events))
	for _, e := range events {
		result = append(result, &port.DLQEvent{
			EventID:      e.EventID,
			EventType:    e.EventType,
			Payload:      e.Payload,
			Status:       e.Status,
			ErrorMessage: e.ErrorMessage,
			CreatedAt:    e.CreatedAt,
			ReplayCount:  e.ReplayCount,
		})
	}
	return result, nil
}

func (a *dlqControllerAdapter) GetDeadEvent(ctx context.Context, eventID uuid.UUID) (*port.DLQEvent, error) {
	e, err := a.delegate.GetDeadEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return &port.DLQEvent{
		EventID:      e.EventID,
		EventType:    e.EventType,
		Payload:      e.Payload,
		Status:       e.Status,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt,
		ReplayCount:  e.ReplayCount,
	}, nil
}

func (a *dlqControllerAdapter) ReDriveEvent(ctx context.Context, eventID uuid.UUID, operatorID uuid.UUID) error {
	return a.delegate.ReDriveEvent(ctx, eventID, operatorID)
}

func (a *dlqControllerAdapter) ReDriveBatch(ctx context.Context, eventIDs []uuid.UUID, operatorID uuid.UUID) (int, error) {
	return a.delegate.ReDriveBatch(ctx, eventIDs, operatorID)
}

type DLQEvent = port.DLQEvent
