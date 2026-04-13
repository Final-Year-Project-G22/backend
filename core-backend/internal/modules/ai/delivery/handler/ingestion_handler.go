package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
)

type IngestionHandler struct {
	ingestionService *service.IngestionService
}

func NewIngestionHandler(ingestionService *service.IngestionService) *IngestionHandler {
	return &IngestionHandler{ingestionService: ingestionService}
}

func (h *IngestionHandler) Healthcheck(ctx context.Context) error {
	return h.ingestionService.Ping(ctx)
}
