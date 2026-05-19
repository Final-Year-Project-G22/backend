package service

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"gorm.io/datatypes"
)

const userNotifPollInterval = 10 * time.Second
const userNotifBatchSize = 50

type UserNotificationScheduler struct {
	scheduledRepo notifrepo.UserScheduledNotificationRepository
	queueRepo     notifrepo.NotificationQueueRepository
	accountReader notifrepo.AccountReader
	logger        core.Logger
}

func NewUserNotificationScheduler(
	scheduledRepo notifrepo.UserScheduledNotificationRepository,
	queueRepo notifrepo.NotificationQueueRepository,
	accountReader notifrepo.AccountReader,
	logger core.Logger,
) *UserNotificationScheduler {
	return &UserNotificationScheduler{
		scheduledRepo: scheduledRepo,
		queueRepo:     queueRepo,
		accountReader: accountReader,
		logger:        logger,
	}
}

func (s *UserNotificationScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *UserNotificationScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(userNotifPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processDue(ctx)
		}
	}
}

func (s *UserNotificationScheduler) processDue(ctx context.Context) {
	due, err := s.scheduledRepo.FetchDue(ctx, userNotifBatchSize)
	if err != nil {
		s.logger.Error("Failed to fetch due scheduled notifications", core.Error(err))
		return
	}

	for _, notif := range due {
		if err := s.enqueueAll(ctx, notif); err != nil {
			s.logger.Error("Failed to enqueue scheduled notification",
				core.String("scheduledID", notif.ID.String()),
				core.Error(err),
			)
			continue
		}
		if err := s.scheduledRepo.MarkSent(ctx, notif.ID); err != nil {
			s.logger.Error("Failed to mark scheduled notification as sent",
				core.String("scheduledID", notif.ID.String()),
				core.Error(err),
			)
		}
	}
}

func (s *UserNotificationScheduler) enqueueAll(ctx context.Context, notif *entity.UserScheduledNotification) error {
	accountInfo, err := s.accountReader.GetAccountInfo(ctx, notif.AccountID)
	if err != nil {
		return err
	}

	for _, ch := range notif.Channels {
		channel := entity.Channel(ch)
		if err := s.enqueue(ctx, notif, channel, accountInfo.Email); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserNotificationScheduler) enqueue(ctx context.Context, notif *entity.UserScheduledNotification, channel entity.Channel, email string) error {
	payload := datatypes.JSONMap{
		"title":   notif.Title,
		"body":    notif.Body,
		"content": notif.Body,
		"subject": notif.Title,
	}

	if channel == entity.ChannelEmail {
		payload["to"] = email
	}

	queueItem := &entity.NotificationQueue{
		NotificationType: entity.NotificationTypeUserScheduled,
		AccountID:        notif.AccountID,
		Priority:         entity.NotificationPriorityMedium,
		Channel:          channel,
		Payload:          payload,
		ScheduledFor:     time.Now().UTC(),
		MaxRetries:       3,
		RetryCount:       0,
		Status:           entity.NotificationStatusPending,
	}

	return s.queueRepo.Create(ctx, queueItem)
}
