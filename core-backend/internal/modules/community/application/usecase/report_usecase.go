package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type contentReportUsecase struct {
	reportRepo repository.ContentReportRepository
	threadRepo repository.DiscussionThreadRepository
	postRepo   repository.DiscussionPostRepository
	blockRepo  repository.ThreadBlockedUserRepository
	transactor sharedrepo.Transactor
}

func NewContentReportUsecase(
	reportRepo repository.ContentReportRepository,
	threadRepo repository.DiscussionThreadRepository,
	postRepo repository.DiscussionPostRepository,
	blockRepo repository.ThreadBlockedUserRepository,
	transactor sharedrepo.Transactor,
) usecase.ContentReportUsecase {
	return &contentReportUsecase{
		reportRepo: reportRepo,
		threadRepo: threadRepo,
		postRepo:   postRepo,
		blockRepo:  blockRepo,
		transactor: transactor,
	}
}

func (u *contentReportUsecase) ReportThread(ctx context.Context, reporterID uuid.UUID, input usecase.ReportThreadInput) (*entity.ContentReport, error) {
	if input.ThreadID == uuid.Nil {
		return nil, apperrors.RequiredFieldError("threadId")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, apperrors.RequiredFieldError("reason")
	}

	if _, err := u.threadRepo.GetByID(ctx, input.ThreadID); err != nil {
		return nil, err
	}

	report := &entity.ContentReport{
		ReporterAccountID: reporterID,
		ThreadID:          &input.ThreadID,
		Reason:            strings.TrimSpace(input.Reason),
		Status:            entity.ReportStatusPending,
	}
	if err := u.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (u *contentReportUsecase) ReportPost(ctx context.Context, reporterID uuid.UUID, input usecase.ReportPostInput) (*entity.ContentReport, error) {
	if input.PostID == uuid.Nil {
		return nil, apperrors.RequiredFieldError("postId")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, apperrors.RequiredFieldError("reason")
	}

	post, err := u.postRepo.GetByID(ctx, input.PostID)
	if err != nil {
		return nil, err
	}

	report := &entity.ContentReport{
		ReporterAccountID: reporterID,
		ThreadID:          &input.ThreadID,
		PostID:            &input.PostID,
		Reason:            strings.TrimSpace(input.Reason),
		Status:            entity.ReportStatusPending,
	}
	if post.ThreadID != input.ThreadID {
		return nil, apperrors.InvalidInputError("postId", "community.errors.postNotInThread")
	}
	if err := u.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (u *contentReportUsecase) ReportUser(ctx context.Context, reporterID uuid.UUID, input usecase.ReportUserInput) (*entity.ContentReport, error) {
	if input.ThreadID == uuid.Nil {
		return nil, apperrors.RequiredFieldError("threadId")
	}
	if input.ReportedAccountID == uuid.Nil {
		return nil, apperrors.RequiredFieldError("reportedAccountId")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, apperrors.RequiredFieldError("reason")
	}

	if _, err := u.threadRepo.GetByID(ctx, input.ThreadID); err != nil {
		return nil, err
	}

	report := &entity.ContentReport{
		ReporterAccountID: reporterID,
		ThreadID:          &input.ThreadID,
		ReportedAccountID: &input.ReportedAccountID,
		Reason:            strings.TrimSpace(input.Reason),
		Status:            entity.ReportStatusPending,
	}
	if err := u.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (u *contentReportUsecase) GetThreadReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error) {
	return u.getReport(ctx, reportID)
}

func (u *contentReportUsecase) GetPostReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error) {
	return u.getReport(ctx, reportID)
}

func (u *contentReportUsecase) GetUserReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error) {
	return u.getReport(ctx, reportID)
}

func (u *contentReportUsecase) getReport(ctx context.Context, reportID uuid.UUID) (*entity.ContentReport, error) {
	report, err := u.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if err == communityerror.ErrReportNotFound {
			return nil, apperrors.NotFoundError("report", reportID.String())
		}
		return nil, err
	}
	return report, nil
}

func (u *contentReportUsecase) ListThreadReports(ctx context.Context, status *entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	if status != nil {
		return u.reportRepo.ListByThreadStatus(ctx, *status, q)
	}
	return u.reportRepo.ListAllThreadReports(ctx, q)
}

func (u *contentReportUsecase) ListPostReports(ctx context.Context, status *entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	if status != nil {
		return u.reportRepo.ListByPostStatus(ctx, *status, q)
	}
	return u.reportRepo.ListAllPostReports(ctx, q)
}

func (u *contentReportUsecase) ListUserReports(ctx context.Context, status *entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	if status != nil {
		return u.reportRepo.ListByUserStatus(ctx, *status, q)
	}
	return u.reportRepo.ListAllUserReports(ctx, q)
}

func (u *contentReportUsecase) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, input usecase.UpdateReportStatusInput, resolvedBy uuid.UUID) (*entity.ContentReport, error) {
	_, err := u.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if err == communityerror.ErrReportNotFound {
			return nil, apperrors.NotFoundError("report", reportID.String())
		}
		return nil, err
	}

	now := time.Now().UTC()
	adminNote := input.AdminNote
	if adminNote != nil {
		note := strings.TrimSpace(*adminNote)
		if note == "" {
			adminNote = nil
		} else {
			adminNote = &note
		}
	}

	err = u.reportRepo.UpdateStatus(ctx, reportID, input.Status, adminNote, &resolvedBy, &now)
	if err != nil {
		return nil, err
	}

	return u.reportRepo.GetByID(ctx, reportID)
}

func (u *contentReportUsecase) DeleteReportedThread(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID) error {
	report, err := u.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if err == communityerror.ErrReportNotFound {
			return apperrors.NotFoundError("report", reportID.String())
		}
		return err
	}

	if report.ThreadID == nil {
		return apperrors.InvalidInputError("reportId", "community.errors.invalidReportType")
	}

	thread, err := u.threadRepo.GetByID(ctx, *report.ThreadID)
	if err != nil {
		return err
	}

	posts, err := u.postRepo.ListByThread(ctx, thread.ID, query.QueryOptions{PageSize: 10000})
	if err != nil {
		return err
	}

	for _, post := range posts {
		if err := u.postRepo.HardDelete(ctx, post.ID); err != nil {
			return err
		}
	}

	if err := u.threadRepo.HardDelete(ctx, thread.ID); err != nil {
		return err
	}

	now := time.Now().UTC()
	return u.reportRepo.UpdateStatus(ctx, reportID, entity.ReportStatusResolved, nil, &adminID, &now)
}

func (u *contentReportUsecase) DeleteReportedPost(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID) error {
	report, err := u.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if err == communityerror.ErrReportNotFound {
			return apperrors.NotFoundError("report", reportID.String())
		}
		return err
	}

	if report.PostID == nil {
		return apperrors.InvalidInputError("reportId", "community.errors.invalidReportType")
	}

	post, err := u.postRepo.GetByID(ctx, *report.PostID)
	if err != nil {
		return err
	}
	threadID := post.ThreadID

	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.postRepo.HardDelete(txCtx, *report.PostID); err != nil {
			return err
		}
		return u.threadRepo.UpdateReplyCount(txCtx, threadID, -1)
	}); err != nil {
		return err
	}

	now := time.Now().UTC()
	return u.reportRepo.UpdateStatus(ctx, reportID, entity.ReportStatusResolved, nil, &adminID, &now)
}

func (u *contentReportUsecase) BlockReportedUser(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID, reason *string) error {
	report, err := u.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		if err == communityerror.ErrReportNotFound {
			return apperrors.NotFoundError("report", reportID.String())
		}
		return err
	}

	if report.ReportedAccountID == nil || report.ThreadID == nil {
		return apperrors.InvalidInputError("reportId", "community.errors.invalidReportType")
	}

	if _, err := u.threadRepo.GetByID(ctx, *report.ThreadID); err != nil {
		return apperrors.NotFoundError("thread", report.ThreadID.String())
	}

	if err := u.blockRepo.Block(ctx, *report.ReportedAccountID, *report.ThreadID, adminID, reason); err != nil {
		return err
	}

	now := time.Now().UTC()
	return u.reportRepo.UpdateStatus(ctx, reportID, entity.ReportStatusResolved, nil, &adminID, &now)
}

func (u *contentReportUsecase) ResolveUserReport(ctx context.Context, threadID, accountID, resolvedBy uuid.UUID) error {
	report, err := u.reportRepo.FindByThreadAndAccount(ctx, threadID, accountID)
	if err != nil {
		return err
	}
	if report == nil {
		return nil
	}

	now := time.Now().UTC()
	return u.reportRepo.UpdateStatus(ctx, report.ID, entity.ReportStatusResolved, nil, &resolvedBy, &now)
}
