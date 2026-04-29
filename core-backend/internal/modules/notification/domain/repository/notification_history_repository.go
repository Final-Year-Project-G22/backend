package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationHistoryRepository interface {
	sharedrepo.GenericRepository[entity.NotificationHistory]

	ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.NotificationHistory, error)
	UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status entity.DeliveryStatus, deliveredAt *time.Time) error
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkClicked(ctx context.Context, id uuid.UUID) error
}
