package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ContentReportRepository interface {
	sharedrepo.GenericRepository[entity.ContentReport]

	ListByTarget(ctx context.Context, targetType entity.TargetType, targetID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, error)
	ListByStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, error)
	UpdateStatus(ctx context.Context, reportID uuid.UUID, status entity.ReportStatus, adminNote *string, resolvedBy *uuid.UUID, resolvedAt *time.Time) error
}
