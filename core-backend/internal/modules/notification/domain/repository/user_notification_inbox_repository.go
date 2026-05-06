package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type UserNotificationInboxRepository interface {
	sharedrepo.GenericRepository[entity.UserNotificationInbox]

	ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserNotificationInbox, int64, error)
	GetUnreadCount(ctx context.Context, accountID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, accountID uuid.UUID) error
	Archive(ctx context.Context, id uuid.UUID) error
	ExpireOld(ctx context.Context, before time.Time) error
	MarkAllReadByCategory(ctx context.Context, accountID uuid.UUID, category entity.NotificationCategory) error
}
