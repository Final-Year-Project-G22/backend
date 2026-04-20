package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type ToggleHandler struct {
	ingestControl port.IngestControl
}

func NewToggleHandler(ingestControl port.IngestControl) *ToggleHandler {
	return &ToggleHandler{ingestControl: ingestControl}
}

func (h *ToggleHandler) HandleGetIngestToggle(ctx context.Context, _ *dto.GetIngestToggleInput) (*dto.GetIngestToggleOutput, error) {
	enabled, _, err := h.ingestControl.GetToggleState(ctx)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetIngestToggleOutput{Body: dto.IngestToggleStateResponse{Enabled: enabled}}, nil
}

func (h *ToggleHandler) HandleSetIngestToggle(ctx context.Context, input *dto.SetIngestToggleInput) (*dto.SetIngestToggleOutput, error) {
	if err := h.ingestControl.SetEnabled(ctx, input.Body.Enabled); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.SetIngestToggleOutput{Body: dto.IngestToggleStateResponse{Enabled: input.Body.Enabled}}, nil
}
