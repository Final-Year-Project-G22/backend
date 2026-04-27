package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type notificationDeliveryUsecase struct {
	queueRepo   repository.NotificationQueueRepository
	historyRepo repository.NotificationHistoryRepository
	inboxRepo   repository.UserNotificationInboxRepository
	transactor  sharedrepo.Transactor
}

func NewNotificationDeliveryUsecase(
	queueRepo repository.NotificationQueueRepository,
	historyRepo repository.NotificationHistoryRepository,
	inboxRepo repository.UserNotificationInboxRepository,
	transactor sharedrepo.Transactor,
) usecase.NotificationDeliveryUsecase {
	return &notificationDeliveryUsecase{
		queueRepo:   queueRepo,
		historyRepo: historyRepo,
		inboxRepo:   inboxRepo,
		transactor:  transactor,
	}
}

func (uc *notificationDeliveryUsecase) ProcessQueue(ctx context.Context, batchSize int) error {
	items, err := uc.queueRepo.FetchPending(ctx, batchSize)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := uc.DeliverItem(ctx, item.ID); err != nil {
			continue
		}
	}
	return nil
}

func (uc *notificationDeliveryUsecase) DeliverItem(ctx context.Context, queueID uuid.UUID) error {
	item, err := uc.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return err
	}
	if item.Status != entity.NotificationStatusPending {
		return nil
	}
	if err := uc.queueRepo.MarkProcessing(ctx, queueID); err != nil {
		return err
	}

	switch item.Channel {
	case entity.ChannelInApp:
		return uc.deliverInApp(ctx, item)
	default:
		return nil
	}
}

func (uc *notificationDeliveryUsecase) HandleDeliveryResult(ctx context.Context, queueID uuid.UUID, success bool, errMsg *string) error {
	if success {
		item, err := uc.queueRepo.GetByID(ctx, queueID)
		if err != nil {
			return err
		}
		if err := uc.createHistoryAndInbox(ctx, item); err != nil {
			msg := err.Error()
			return uc.queueRepo.MarkFailed(ctx, queueID, msg)
		}
		return uc.queueRepo.MarkDelivered(ctx, queueID)
	}

	item, err := uc.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return err
	}

	if item.RetryCount < item.MaxRetries {
		backoff := retryBackoff(item.RetryCount)
		return uc.queueRepo.IncrementRetry(ctx, queueID, time.Now().UTC().Add(backoff))
	}

	msg := ""
	if errMsg != nil {
		msg = *errMsg
	}
	return uc.queueRepo.MarkFailed(ctx, queueID, msg)
}

func (uc *notificationDeliveryUsecase) RetryFailed(ctx context.Context, batchSize int) error {
	opts := query.DefaultQueryOptions()
	opts.Filters = map[string]interface{}{
		"status": entity.NotificationStatusFailed,
	}
	opts.PageSize = batchSize

	items, err := uc.queueRepo.Find(ctx, opts)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.RetryCount >= item.MaxRetries {
			continue
		}
		backoff := retryBackoff(item.RetryCount)
		updates := map[string]interface{}{
			"status":        entity.NotificationStatusPending,
			"scheduled_for": time.Now().UTC().Add(backoff),
		}
		if err := uc.queueRepo.UpdateByID(ctx, item.ID, updates); err != nil {
			return err
		}
	}

	return nil
}

func (uc *notificationDeliveryUsecase) CancelPendingForAccount(ctx context.Context, accountID uuid.UUID) error {
	return uc.queueRepo.CancelByAccount(ctx, accountID)
}

func (uc *notificationDeliveryUsecase) deliverInApp(ctx context.Context, item *entity.NotificationQueue) error {
	return uc.HandleDeliveryResult(ctx, item.ID, true, nil)
}

func (uc *notificationDeliveryUsecase) createHistoryAndInbox(ctx context.Context, item *entity.NotificationQueue) error {
	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		title, _ := item.Payload["title"].(string)
		content, _ := item.Payload["content"].(string)
		actionUrlStr, _ := item.Payload["actionUrl"].(string)

		var actionUrl *string
		if actionUrlStr != "" {
			actionUrl = &actionUrlStr
		}

		now := time.Now().UTC()

		history := &entity.NotificationHistory{
			AccountID:        item.AccountID,
			NotificationType: item.NotificationType,
			Channel:          item.Channel,
			Title:            title,
			Content:          content,
			ActionUrl:        actionUrl,
			SentAt:           now,
			DeliveredAt:      &now,
			DeliveryStatus:   entity.DeliveryStatusDelivered,
		}

		if err := uc.historyRepo.Create(txCtx, history); err != nil {
			return fmt.Errorf("failed to create history entry: %w", err)
		}

		inbox := &entity.UserNotificationInbox{
			AccountID:             item.AccountID,
			NotificationHistoryID: history.ID,
			Category:              uc.toCategory(item.NotificationType),
			ActionUrl:             actionUrl,
			IsRead:                false,
			IsArchived:            false,
		}

		if err := uc.inboxRepo.Create(txCtx, inbox); err != nil {
			return fmt.Errorf("failed to create inbox entry: %w", err)
		}

		return nil
	})
}

func (uc *notificationDeliveryUsecase) toCategory(nt entity.NotificationType) entity.NotificationCategory {
	switch nt {
	case entity.NotificationTypeSystemAnnouncement, entity.NotificationTypePolicyUpdate, entity.NotificationTypeWelcomeMessage:
		return entity.NotificationCategorySystem
	case entity.NotificationTypeCommunityReply, entity.NotificationTypeCommunitySolution, entity.NotificationTypeCommunityMention:
		return entity.NotificationCategoryCommunity
	case entity.NotificationTypeGuideStepCompleted, entity.NotificationTypeGuideDeadline, entity.NotificationTypeGuideUpdate:
		return entity.NotificationCategoryGuide
	case entity.NotificationTypeAIQuotaLimit, entity.NotificationTypeAIResponseReady:
		return entity.NotificationCategoryAI
	case entity.NotificationTypeAccountAlert, entity.NotificationTypeAccountVerification, entity.NotificationTypePasswordReset:
		return entity.NotificationCategorySecurity
	case entity.NotificationTypePaymentConfirmation:
		return entity.NotificationCategoryPayment
	default:
		return entity.NotificationCategorySystem
	}
}

func retryBackoff(retryCount int) time.Duration {
	switch retryCount {
	case 0:
		return 1 * time.Minute
	case 1:
		return 2 * time.Minute
	case 2:
		return 4 * time.Minute
	default:
		return 8 * time.Minute
	}
}
