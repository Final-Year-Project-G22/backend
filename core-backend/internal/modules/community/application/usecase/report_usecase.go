package usecase

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type contentReportUsecase struct {
	reportRepo repository.ContentReportRepository
	threadRepo repository.DiscussionThreadRepository
	postRepo   repository.DiscussionPostRepository
}

func NewContentReportUsecase(
	reportRepo repository.ContentReportRepository,
	threadRepo repository.DiscussionThreadRepository,
	postRepo repository.DiscussionPostRepository,
) usecase.ContentReportUsecase {
	return &contentReportUsecase{
		reportRepo: reportRepo,
		threadRepo: threadRepo,
		postRepo:   postRepo,
	}
}

func (u *contentReportUsecase) ReportContent(ctx context.Context, reporterID uuid.UUID, input usecase.ReportInput) (*entity.ContentReport, error) {
	if input.TargetID == uuid.Nil {
		return nil, apperrors.RequiredFieldError("targetId")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, apperrors.RequiredFieldError("reason")
	}

	switch input.TargetType {
	case entity.TargetTypeThread:
		if _, err := u.threadRepo.GetByID(ctx, input.TargetID); err != nil {
			return nil, err
		}
	case entity.TargetTypePost:
		if _, err := u.postRepo.GetByID(ctx, input.TargetID); err != nil {
			return nil, err
		}
	case entity.TargetTypeUser, entity.TargetTypeHidden:
		// user existence validation handled in IAM layer when needed
	default:
		return nil, apperrors.InvalidInputError("targetType", "community.errors.invalidTargetType")
	}

	report := &entity.ContentReport{
		ReporterAccountID: reporterID,
		TargetType:        input.TargetType,
		TargetID:          input.TargetID,
		Reason:            strings.TrimSpace(input.Reason),
		Status:            entity.ReportStatusPending,
	}
	if err := u.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}
