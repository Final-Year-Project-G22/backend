package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationHistoryUsecase interface {
	ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.NotificationHistory, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationHistory, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkClicked(ctx context.Context, id uuid.UUID) error
	UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status entity.DeliveryStatus, deliveredAt *time.Time) error
}
