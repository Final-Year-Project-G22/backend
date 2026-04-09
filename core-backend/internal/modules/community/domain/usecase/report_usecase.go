package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/google/uuid"
)

type ContentReportUsecase interface {
	ReportContent(ctx context.Context, reporterID uuid.UUID, input ReportInput) (*entity.ContentReport, error)
}
