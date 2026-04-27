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

	CountAll(ctx context.Context) (int64, error)
	CountByThreadID(ctx context.Context, threadID uuid.UUID) (int64, error)
	CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
	CountByReportedAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)

	FindByThreadAndAccount(ctx context.Context, threadID, accountID uuid.UUID) (*entity.ContentReport, error)

	ListByThreadID(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByThreadIDAndStatus(ctx context.Context, threadID uuid.UUID, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByPostID(ctx context.Context, postID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByPostIDAndStatus(ctx context.Context, postID uuid.UUID, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByReportedAccountID(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByReportedAccountIDAndStatus(ctx context.Context, accountID uuid.UUID, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)

	ListAllThreadReports(ctx context.Context, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListAllPostReports(ctx context.Context, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListAllUserReports(ctx context.Context, q query.QueryOptions) ([]*entity.ContentReport, int64, error)

	ListByThreadStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByPostStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListByUserStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)

	UpdateStatus(ctx context.Context, reportID uuid.UUID, status entity.ReportStatus, adminNote *string, resolvedBy *uuid.UUID, resolvedAt *time.Time) error
}
