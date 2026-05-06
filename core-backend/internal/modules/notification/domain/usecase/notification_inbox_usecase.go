package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationInboxUsecase interface {
	ListInbox(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserNotificationInbox, int64, error)
	GetUnreadCount(ctx context.Context, accountID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, accountID uuid.UUID, inboxID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, accountID uuid.UUID) error
	MarkCategoryAsRead(ctx context.Context, accountID uuid.UUID, category entity.NotificationCategory) error
	ArchiveNotification(ctx context.Context, accountID uuid.UUID, inboxID uuid.UUID) error
	DeleteNotification(ctx context.Context, accountID uuid.UUID, inboxID uuid.UUID) error
	ExpireOld(ctx context.Context, before time.Time) error
}
