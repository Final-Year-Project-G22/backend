package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type notificationDeliveryUsecase struct {
	queueRepo       repository.NotificationQueueRepository
	historyRepo     repository.NotificationHistoryRepository
	inboxRepo       repository.UserNotificationInboxRepository
	deliveryLogRepo repository.EmailDeliveryLogRepository
	emailProvider   repository.EmailProvider
	deviceRepo      repository.UserDeviceRepository
	pushProvider    repository.PushProvider
	sseBroadcaster  *service.InboxSSEBroadcaster
	transactor      sharedrepo.Transactor
	logger          core.Logger
	cfg             *core.Config
}

func NewNotificationDeliveryUsecase(
	queueRepo repository.NotificationQueueRepository,
	historyRepo repository.NotificationHistoryRepository,
	inboxRepo repository.UserNotificationInboxRepository,
	deliveryLogRepo repository.EmailDeliveryLogRepository,
	emailProvider repository.EmailProvider,
	deviceRepo repository.UserDeviceRepository,
	pushProvider repository.PushProvider,
	sseBroadcaster *service.InboxSSEBroadcaster,
	transactor sharedrepo.Transactor,
	logger core.Logger,
	cfg *core.Config,
) usecase.NotificationDeliveryUsecase {
	return &notificationDeliveryUsecase{
		queueRepo:       queueRepo,
		historyRepo:     historyRepo,
		inboxRepo:       inboxRepo,
		deliveryLogRepo: deliveryLogRepo,
		emailProvider:   emailProvider,
		deviceRepo:      deviceRepo,
		pushProvider:    pushProvider,
		sseBroadcaster:  sseBroadcaster,
		transactor:      transactor,
		logger:          logger,
		cfg:             cfg,
	}
}

func (uc *notificationDeliveryUsecase) ProcessQueue(ctx context.Context, batchSize int) error {
	items, err := uc.queueRepo.FetchPending(ctx, batchSize)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := uc.DeliverItem(ctx, item.ID); err != nil {
			uc.logger.Error("Failed to deliver queue item",
				core.String("queueID", item.ID.String()),
				core.Error(err),
			)
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
	case entity.ChannelEmail:
		return uc.deliverEmail(ctx, item)
	case entity.ChannelPush:
		return uc.deliverPush(ctx, item)
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

func (uc *notificationDeliveryUsecase) deliverEmail(ctx context.Context, item *entity.NotificationQueue) error {
	to, _ := item.Payload["to"].(string)
	subject, _ := item.Payload["subject"].(string)
	body, _ := item.Payload["body"].(string)

	if to == "" || subject == "" {
		msg := "missing required email payload fields"
		return uc.HandleDeliveryResult(ctx, item.ID, false, &msg)
	}

	metadata := map[string]string{
		"X-Notification-ID":   item.ID.String(),
		"X-Notification-Type": string(item.NotificationType),
	}

	providerMsgID, err := uc.emailProvider.Send(ctx, to, subject, body, metadata)
	if err != nil {
		msg := err.Error()
		uc.logger.Error("Email send failed",
			core.String("queueID", item.ID.String()),
			core.String("to", to),
			core.Error(err),
		)
		return uc.HandleDeliveryResult(ctx, item.ID, false, &msg)
	}

	return uc.createHistoryInboxAndDeliveryLog(ctx, item, providerMsgID, to, subject)
}

func (uc *notificationDeliveryUsecase) deliverInApp(ctx context.Context, item *entity.NotificationQueue) error {
	return uc.HandleDeliveryResult(ctx, item.ID, true, nil)
}

func (uc *notificationDeliveryUsecase) deliverPush(ctx context.Context, item *entity.NotificationQueue) error {
	title, _ := item.Payload["title"].(string)
	body, _ := item.Payload["body"].(string)

	devices, err := uc.deviceRepo.ListByAccount(ctx, item.AccountID)
	if err != nil {
		msg := err.Error()
		return uc.HandleDeliveryResult(ctx, item.ID, false, &msg)
	}

	seen := make(map[string]bool)
	for _, dev := range devices {
		if dev.PushToken == nil || *dev.PushToken == "" || seen[*dev.PushToken] {
			continue
		}
		seen[*dev.PushToken] = true
		data := map[string]string{
			"notificationType": string(item.NotificationType),
			"queueID":          item.ID.String(),
		}
		if err := uc.pushProvider.Send(ctx, *dev.PushToken, title, body, data); err != nil {
			uc.logger.Error("Push send failed", core.String("token", *dev.PushToken), core.Error(err))
		}
	}

	return uc.createHistoryForPush(ctx, item)
}

func (uc *notificationDeliveryUsecase) createHistoryForPush(ctx context.Context, item *entity.NotificationQueue) error {
	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		title, _ := item.Payload["title"].(string)
		content, _ := item.Payload["body"].(string)
		now := time.Now().UTC()

		history := &entity.NotificationHistory{
			AccountID:        item.AccountID,
			NotificationType: item.NotificationType,
			Channel:          item.Channel,
			Title:            title,
			Content:          content,
			SentAt:           now,
			DeliveredAt:      &now,
			DeliveryStatus:   entity.DeliveryStatusDelivered,
		}
		if err := uc.historyRepo.Create(txCtx, history); err != nil {
			return fmt.Errorf("failed to create history entry: %w", err)
		}
		return uc.queueRepo.MarkDelivered(txCtx, item.ID)
	})
}

func (uc *notificationDeliveryUsecase) createHistoryInboxAndDeliveryLog(ctx context.Context, item *entity.NotificationQueue, providerMsgID, to, subject string) error {
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		title, _ := item.Payload["title"].(string)
		content, _ := item.Payload["content"].(string)
		actionUrlStr, _ := item.Payload["actionUrl"].(string)
		isMuted, _ := item.Payload["_isMuted"].(bool)

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

		providerName := "resend"
		if uc.cfg.Email.Enabled {
			providerName = "smtp"
		}

		deliveryLog := &entity.EmailDeliveryLog{
			NotificationHistoryID: history.ID,
			Provider:              providerName,
			ProviderMessageID:     &providerMsgID,
			RecipientEmail:        to,
			Subject:               subject,
			SentAt:                now,
			DeliveryStatus:        entity.DeliveryStatusSent,
		}

		if err := uc.deliveryLogRepo.Create(txCtx, deliveryLog); err != nil {
			return fmt.Errorf("failed to create email delivery log: %w", err)
		}

		if isMuted || item.Channel != entity.ChannelInApp {
			return uc.queueRepo.MarkDelivered(txCtx, item.ID)
		}

		inbox := &entity.UserNotificationInbox{
			AccountID:             item.AccountID,
			NotificationHistoryID: history.ID,
			ActionUrl:             actionUrl,
			IsRead:                false,
			IsArchived:            false,
		}

		if err := uc.inboxRepo.Create(txCtx, inbox); err != nil {
			return fmt.Errorf("failed to create inbox entry: %w", err)
		}

		return uc.queueRepo.MarkDelivered(txCtx, item.ID)
	})
	if err != nil {
		return err
	}

	if item.Channel == entity.ChannelInApp {
		unreadCount, _ := uc.inboxRepo.GetUnreadCount(ctx, item.AccountID)
		uc.sseBroadcaster.Publish(service.InboxEvent{
			Type:             "notification_new",
			AccountID:        item.AccountID,
			NotificationType: item.NotificationType,
			UnreadCount:      unreadCount,
			Timestamp:        time.Now().UTC(),
		})
	}
	return nil
}

func (uc *notificationDeliveryUsecase) createHistoryAndInbox(ctx context.Context, item *entity.NotificationQueue) error {
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		title, _ := item.Payload["title"].(string)
		content, _ := item.Payload["content"].(string)
		actionUrlStr, _ := item.Payload["actionUrl"].(string)
		isMuted, _ := item.Payload["_isMuted"].(bool)

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

		if isMuted {
			return nil
		}

		inbox := &entity.UserNotificationInbox{
			AccountID:             item.AccountID,
			NotificationHistoryID: history.ID,
			ActionUrl:             actionUrl,
			IsRead:                false,
			IsArchived:            false,
		}

		if err := uc.inboxRepo.Create(txCtx, inbox); err != nil {
			return fmt.Errorf("failed to create inbox entry: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	unreadCount, _ := uc.inboxRepo.GetUnreadCount(ctx, item.AccountID)
	uc.sseBroadcaster.Publish(service.InboxEvent{
		Type:             "notification_new",
		AccountID:        item.AccountID,
		NotificationType: item.NotificationType,
		UnreadCount:      unreadCount,
		Timestamp:        time.Now().UTC(),
	})
	return nil
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
