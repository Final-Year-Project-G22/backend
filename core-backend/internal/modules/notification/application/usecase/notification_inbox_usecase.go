package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type notificationInboxUsecase struct {
	inboxRepo   repository.UserNotificationInboxRepository
	historyRepo repository.NotificationHistoryRepository
}

func NewNotificationInboxUsecase(
	inboxRepo repository.UserNotificationInboxRepository,
	historyRepo repository.NotificationHistoryRepository,
) usecase.NotificationInboxUsecase {
	return &notificationInboxUsecase{
		inboxRepo:   inboxRepo,
		historyRepo: historyRepo,
	}
}

func (uc *notificationInboxUsecase) ListInbox(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserNotificationInbox, int64, error) {
	if q.Preload == nil {
		q.Preload = []string{"NotificationHistory"}
	} else {
		q.Preload = append(q.Preload, "NotificationHistory")
	}
	return uc.inboxRepo.ListByAccount(ctx, accountID, q)
}

func (uc *notificationInboxUsecase) GetUnreadCount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	return uc.inboxRepo.GetUnreadCount(ctx, accountID)
}

func (uc *notificationInboxUsecase) MarkAsRead(ctx context.Context, accountID uuid.UUID, inboxID uuid.UUID) error {
	inbox, err := uc.inboxRepo.GetByID(ctx, inboxID)
	if err != nil {
		return err
	}
	if inbox.AccountID != accountID {
		return notiferror.ErrInboxEntryNotFound
	}

	if err := uc.inboxRepo.MarkAsRead(ctx, inboxID); err != nil {
		return err
	}

	return uc.historyRepo.MarkRead(ctx, inbox.NotificationHistoryID)
}

func (uc *notificationInboxUsecase) MarkAllAsRead(ctx context.Context, accountID uuid.UUID) error {
	return uc.inboxRepo.MarkAllAsRead(ctx, accountID)
}

func (uc *notificationInboxUsecase) MarkCategoryAsRead(ctx context.Context, accountID uuid.UUID, category entity.NotificationCategory) error {
	return uc.inboxRepo.MarkAllReadByCategory(ctx, accountID, category)
}

func (uc *notificationInboxUsecase) ArchiveNotification(ctx context.Context, accountID uuid.UUID, inboxID uuid.UUID) error {
	inbox, err := uc.inboxRepo.GetByID(ctx, inboxID)
	if err != nil {
		return err
	}
	if inbox.AccountID != accountID {
		return notiferror.ErrInboxEntryNotFound
	}
	return uc.inboxRepo.Archive(ctx, inboxID)
}

func (uc *notificationInboxUsecase) DeleteNotification(ctx context.Context, accountID uuid.UUID, inboxID uuid.UUID) error {
	inbox, err := uc.inboxRepo.GetByID(ctx, inboxID)
	if err != nil {
		return err
	}
	if inbox.AccountID != accountID {
		return notiferror.ErrInboxEntryNotFound
	}
	return uc.inboxRepo.Delete(ctx, inboxID)
}

func (uc *notificationInboxUsecase) ExpireOld(ctx context.Context, before time.Time) error {
	return uc.inboxRepo.ExpireOld(ctx, before)
}
