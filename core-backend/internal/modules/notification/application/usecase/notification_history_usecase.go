package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type notificationHistoryUsecase struct {
	historyRepo repository.NotificationHistoryRepository
}

func NewNotificationHistoryUsecase(
	historyRepo repository.NotificationHistoryRepository,
) usecase.NotificationHistoryUsecase {
	return &notificationHistoryUsecase{
		historyRepo: historyRepo,
	}
}

func (uc *notificationHistoryUsecase) ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.NotificationHistory, error) {
	return uc.historyRepo.ListByAccount(ctx, accountID, q)
}

func (uc *notificationHistoryUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationHistory, error) {
	return uc.historyRepo.GetByID(ctx, id)
}

func (uc *notificationHistoryUsecase) MarkRead(ctx context.Context, id uuid.UUID) error {
	return uc.historyRepo.MarkRead(ctx, id)
}

func (uc *notificationHistoryUsecase) MarkClicked(ctx context.Context, id uuid.UUID) error {
	return uc.historyRepo.MarkClicked(ctx, id)
}

func (uc *notificationHistoryUsecase) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status entity.DeliveryStatus, deliveredAt *time.Time) error {
	return uc.historyRepo.UpdateDeliveryStatus(ctx, id, status, deliveredAt)
}
