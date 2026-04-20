package handler

import (
	"context"
	"encoding/json"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type SSEHandler struct {
	gateway *service.SSEGateway
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
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		payload["eventId"] = eventID
		return send("status", payload)
	})
}
