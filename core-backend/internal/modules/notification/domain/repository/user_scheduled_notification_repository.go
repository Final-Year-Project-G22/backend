package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type UserScheduledNotificationRepository interface {
	sharedrepo.GenericRepository[entity.UserScheduledNotification]

	FetchDue(ctx context.Context, limit int) ([]*entity.UserScheduledNotification, error)
	CountPendingByAccount(ctx context.Context, accountID uuid.UUID) (int64, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.UserScheduledNotification, error)
	MarkSent(ctx context.Context, id uuid.UUID) error
	CancelByID(ctx context.Context, id uuid.UUID) error
}
