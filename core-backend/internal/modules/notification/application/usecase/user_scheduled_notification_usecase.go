package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const maxPendingForNonPro = 3

type userScheduledNotificationUsecase struct {
	repo               notifrepo.UserScheduledNotificationRepository
	tmplRepo           notifrepo.ScheduledAlertTemplateRepository
	subscriptionReader notifrepo.SubscriptionReader
}

func NewUserScheduledNotificationUsecase(
	repo notifrepo.UserScheduledNotificationRepository,
	tmplRepo notifrepo.ScheduledAlertTemplateRepository,
	subscriptionReader notifrepo.SubscriptionReader,
) usecase.UserScheduledNotificationUsecase {
	return &userScheduledNotificationUsecase{
		repo:               repo,
		tmplRepo:           tmplRepo,
		subscriptionReader: subscriptionReader,
	}
}

func (uc *userScheduledNotificationUsecase) Schedule(ctx context.Context, accountID uuid.UUID, input usecase.ScheduleUserNotificationInput) (*entity.UserScheduledNotification, error) {
	if len(input.Channels) == 0 {
		return nil, apperrors.BadRequestError("notification.errors.channelRequired")
	}

	for _, ch := range input.Channels {
		if ch != entity.ChannelInApp && ch != entity.ChannelEmail && ch != entity.ChannelPush {
			return nil, notiferror.ErrInvalidChannel
		}
	}

	if input.Title == "" || input.Body == "" {
		return nil, apperrors.BadRequestError("notification.errors.titleAndBodyRequired")
	}

	if input.ScheduledFor.Before(time.Now().UTC()) {
		return nil, apperrors.BadRequestError("notification.errors.scheduledTimeMustBeFuture")
	}

	count, err := uc.repo.CountPendingByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if count >= maxPendingForNonPro {
		isPro, err := uc.subscriptionReader.HasActiveProSubscription(ctx, accountID)
		if err != nil {
			return nil, err
		}
		if !isPro {
			return nil, notiferror.ErrMaxScheduledAlertsReached
		}
	}

	channels := make(pq.StringArray, len(input.Channels))
	for i, ch := range input.Channels {
		channels[i] = string(ch)
	}

	notif := &entity.UserScheduledNotification{
		AccountID:    accountID,
		TemplateSlug: input.TemplateSlug,
		Title:        input.Title,
		Body:         input.Body,
		Channels:     channels,
		ScheduledFor: input.ScheduledFor,
		Status:       entity.ScheduleStatusPending,
	}

	if err := uc.repo.Create(ctx, notif); err != nil {
		return nil, err
	}
	return notif, nil
}

func (uc *userScheduledNotificationUsecase) List(ctx context.Context, accountID uuid.UUID) ([]*entity.UserScheduledNotification, error) {
	return uc.repo.ListByAccount(ctx, accountID)
}

func (uc *userScheduledNotificationUsecase) GetByID(ctx context.Context, accountID uuid.UUID, id uuid.UUID) (*entity.UserScheduledNotification, error) {
	notif, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if notif.AccountID != accountID {
		return nil, notiferror.ErrScheduledAlertNotFound
	}
	return notif, nil
}

func (uc *userScheduledNotificationUsecase) Cancel(ctx context.Context, accountID uuid.UUID, id uuid.UUID) error {
	notif, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notif.AccountID != accountID {
		return notiferror.ErrScheduledAlertNotFound
	}
	if notif.Status != entity.ScheduleStatusPending {
		return apperrors.BadRequestError("notification.errors.cancelOnlyPending")
	}
	return uc.repo.CancelByID(ctx, id)
}

func (uc *userScheduledNotificationUsecase) Reschedule(ctx context.Context, accountID uuid.UUID, id uuid.UUID, input usecase.RescheduleUserNotificationInput) error {
	notif, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notif.AccountID != accountID {
		return notiferror.ErrScheduledAlertNotFound
	}
	if notif.Status != entity.ScheduleStatusPending {
		return apperrors.BadRequestError("notification.errors.rescheduleOnlyPending")
	}
	if input.ScheduledFor.Before(time.Now().UTC()) {
		return apperrors.BadRequestError("notification.errors.rescheduleTimeMustBeFuture")
	}

	updates := map[string]interface{}{
		"scheduled_for":    input.ScheduledFor,
		"rescheduled_from": notif.ScheduledFor,
		"updated_at":       time.Now().UTC(),
	}
	return uc.repo.UpdateByID(ctx, id, updates)
}

func (uc *userScheduledNotificationUsecase) ListTemplates(ctx context.Context) ([]*entity.ScheduledAlertTemplate, error) {
	return uc.tmplRepo.ListActive(ctx)
}
