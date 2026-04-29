package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ContentReportUsecase interface {
	ReportThread(ctx context.Context, reporterID uuid.UUID, input ReportThreadInput) (*entity.ContentReport, error)
	ReportPost(ctx context.Context, reporterID uuid.UUID, input ReportPostInput) (*entity.ContentReport, error)
	ReportUser(ctx context.Context, reporterID uuid.UUID, input ReportUserInput) (*entity.ContentReport, error)

	GetThreadReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error)
	GetPostReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error)
	GetUserReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error)

	ListThreadReports(ctx context.Context, status *entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListPostReports(ctx context.Context, status *entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)
	ListUserReports(ctx context.Context, status *entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error)

	UpdateReportStatus(ctx context.Context, reportID uuid.UUID, input UpdateReportStatusInput, resolvedBy uuid.UUID) (*entity.ContentReport, error)
	ResolveUserReport(ctx context.Context, threadID, accountID, resolvedBy uuid.UUID) error
	DeleteReportedThread(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID) error
	DeleteReportedPost(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID) error
	BlockReportedUser(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID, reason *string) error
}
