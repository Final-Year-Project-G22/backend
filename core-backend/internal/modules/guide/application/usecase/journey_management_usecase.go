package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/google/uuid"
)

type journeyManagementUsecase struct {
	progressRepo repository.ProgressRepository
}

func NewJourneyManagementUsecase(
	progressRepo repository.ProgressRepository,
) usecase.JourneyManagementUseCase {
	return &journeyManagementUsecase{
		progressRepo: progressRepo,
	}
}

func (u *journeyManagementUsecase) InvalidateUserJourney(ctx context.Context, accountID, userID, guideID uuid.UUID) error {
	return u.progressRepo.InvalidateJourneyForUser(ctx, accountID, userID, guideID)
}

func (u *journeyManagementUsecase) InvalidateAllJourneysForGuide(ctx context.Context, guideID uuid.UUID) error {
	return u.progressRepo.InvalidateJourneysForGuide(ctx, guideID)
}
