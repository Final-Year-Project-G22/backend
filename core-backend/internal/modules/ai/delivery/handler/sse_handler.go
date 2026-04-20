package handler

import (
	"context"
	"encoding/json"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type sseStreamer interface {
	StreamStatusByAccount(ctx context.Context, accountID uuid.UUID, lastEventID string, sendFunc service.SSEDeliveryFunc) error
}

type SSEHandler struct {
	gateway sseStreamer
}

func NewSSEHandler(gateway *service.SSEGateway) *SSEHandler {
	return &SSEHandler{gateway: gateway}
}

func (h *SSEHandler) StreamAccountStatus(ctx context.Context, lastEventID string, send func(event string, payload any) error) error {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ingestion.errors.unauthorized"))
	}

	return h.gateway.StreamStatusByAccount(ctx, accountID, lastEventID, func(eventID string, data []byte) error {
		var pl map[string]any
		if err := json.Unmarshal(data, &pl); err != nil {
			return err
		}
		pl["eventId"] = eventID
		return send("status", pl)
	})
}
