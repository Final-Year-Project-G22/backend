package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type DLQHandler struct {
	dlqController service.DLQController
}

func NewDLQHandler(dlqController service.DLQController) *DLQHandler {
	return &DLQHandler{dlqController: dlqController}
}

func (h *DLQHandler) HandleListDeadEvents(ctx context.Context, input *dto.ListDeadEventsInput) (*dto.ListDeadEventsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, errors.UnauthorizedError("dlq.errors.unauthorized")
	}

	limit := 20
	offset := 0
	if input.Query.Limit != nil && *input.Query.Limit > 0 {
		limit = *input.Query.Limit
	}
	if input.Query.Offset != nil && *input.Query.Offset >= 0 {
		offset = *input.Query.Offset
	}

	events, err := h.dlqController.ListDeadEvents(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.DeadEventDTO, 0, len(events))
	for _, e := range events {
		payload, err := json.Marshal(e.Payload)
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, dto.DeadEventDTO{
			EventID:      e.EventID,
			EventType:    e.EventType,
			Payload:      string(payload),
			Status:       string(e.Status),
			ErrorMessage: e.ErrorMessage,
			CreatedAt:    e.CreatedAt,
			ReplayCount:  e.ReplayCount,
		})
	}

	return &dto.ListDeadEventsOutput{
		Body: struct {
			Events []dto.DeadEventDTO `json:"events"`
			Total  int                `json:"total"`
		}{
			Events: dtos,
			Total:  len(dtos),
		},
	}, nil
}

func (h *DLQHandler) HandleGetDeadEvent(ctx context.Context, input *dto.GetDeadEventInput) (*dto.GetDeadEventOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, errors.UnauthorizedError("dlq.errors.unauthorized")
	}

	eventID, err := parseEventID(ctx)
	if err != nil {
		return nil, errors.BadRequestError("dlq.errors.invalidEventId")
	}

	event, err := h.dlqController.GetDeadEvent(ctx, eventID)
	if err != nil {
		return nil, errors.NotFoundError("event", "event not found")
	}
	if event == nil {
		return nil, errors.NotFoundError("event", "event not found")
	}
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return nil, err
	}

	return &dto.GetDeadEventOutput{
		Body: dto.DeadEventDTO{
			EventID:      event.EventID,
			EventType:    event.EventType,
			Payload:      string(payload),
			Status:       string(event.Status),
			ErrorMessage: event.ErrorMessage,
			CreatedAt:    event.CreatedAt,
			ReplayCount:  event.ReplayCount,
		},
	}, nil
}

func (h *DLQHandler) HandleRedriveEvent(ctx context.Context, input *dto.RedriveEventInput) (*dto.RedriveEventOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	operatorID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
	if accountID == contextkeys.NilUUID || operatorID == contextkeys.NilUUID {
		return nil, errors.UnauthorizedError("dlq.errors.unauthorized")
	}

	eventID, err := parseEventID(ctx)
	if err != nil {
		return nil, errors.BadRequestError("dlq.errors.invalidEventId")
	}

	err = h.dlqController.ReDriveEvent(ctx, eventID, operatorID)
	if err != nil {
		return nil, err
	}

	return &dto.RedriveEventOutput{
		Body: struct {
			Success bool `json:"success"`
		}{Success: true},
	}, nil
}

func (h *DLQHandler) HandleRedriveBatch(ctx context.Context, input *dto.RedriveBatchInput) (*dto.RedriveBatchOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	operatorID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
	if accountID == contextkeys.NilUUID || operatorID == contextkeys.NilUUID {
		return nil, errors.UnauthorizedError("dlq.errors.unauthorized")
	}

	count, err := h.dlqController.ReDriveBatch(ctx, input.Body.EventIDs, operatorID)
	if err != nil {
		return nil, err
	}

	return &dto.RedriveBatchOutput{
		Body: struct {
			SuccessCount int `json:"successCount"`
		}{SuccessCount: count},
	}, nil
}

func parseEventID(ctx context.Context) (uuid.UUID, error) {
	val, ok := ctx.Value("eventId").(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing eventId path param")
	}
	return uuid.Parse(val)
}
