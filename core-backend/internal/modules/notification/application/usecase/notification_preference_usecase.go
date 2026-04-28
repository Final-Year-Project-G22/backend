package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

type IAMGlobalPreferenceReader interface {
	IsNotificationEnabled(ctx context.Context, accountID uuid.UUID) (bool, error)
}

type notificationPreferenceUsecase struct {
	prefRepo  repository.UserNotificationPreferenceRepository
	iamReader IAMGlobalPreferenceReader
}

func NewNotificationPreferenceUsecase(
	prefRepo repository.UserNotificationPreferenceRepository,
	iamReader IAMGlobalPreferenceReader,
) usecase.NotificationPreferenceUsecase {
	return &notificationPreferenceUsecase{
		prefRepo:  prefRepo,
		iamReader: iamReader,
	}
}

func (uc *notificationPreferenceUsecase) SetPreference(ctx context.Context, accountID uuid.UUID, input usecase.SetPreferenceInput) error {
	pref := &entity.UserNotificationPreference{
		AccountID:        accountID,
		NotificationType: input.NotificationType,
		Channel:          input.Channel,
		IsEnabled:        input.IsEnabled,
		QuietHoursStart:  input.QuietHoursStart,
		QuietHoursEnd:    input.QuietHoursEnd,
	}
	return uc.prefRepo.Upsert(ctx, pref)
}

func (uc *notificationPreferenceUsecase) GetPreferences(ctx context.Context, accountID uuid.UUID) ([]*entity.UserNotificationPreference, error) {
	return uc.prefRepo.ListByAccount(ctx, accountID)
}

func (uc *notificationPreferenceUsecase) GetEffectivePreference(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (bool, error) {
	if uc.iamReader != nil {
		enabled, err := uc.iamReader.IsNotificationEnabled(ctx, accountID)
		if err != nil {
			return false, err
		}
		if !enabled {
			return false, nil
		}
	}

	pref, err := uc.prefRepo.GetByAccountAndTypeAndChannel(ctx, accountID, notificationType, channel)
	if err != nil {
		if err == notiferror.ErrPreferenceNotFound {
			return true, nil
		}
		return false, err
	}

	return pref.IsEnabled, nil
}

func (uc *notificationPreferenceUsecase) IsQuietHours(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (bool, error) {
	pref, err := uc.prefRepo.GetByAccountAndTypeAndChannel(ctx, accountID, notificationType, channel)
	if err != nil {
		if err == notiferror.ErrPreferenceNotFound {
			return false, nil
		}
		return false, err
	}

	if pref.QuietHoursStart == nil || pref.QuietHoursEnd == nil {
		return false, nil
	}

	now := time.Now()
	nowTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	start := time.Date(0, 1, 1, pref.QuietHoursStart.Hour(), pref.QuietHoursStart.Minute(), pref.QuietHoursStart.Second(), 0, pref.QuietHoursStart.Location())
	end := time.Date(0, 1, 1, pref.QuietHoursEnd.Hour(), pref.QuietHoursEnd.Minute(), pref.QuietHoursEnd.Second(), 0, pref.QuietHoursEnd.Location())

	return nowTime.After(start) && nowTime.Before(end), nil
}

func (uc *notificationPreferenceUsecase) DeletePreference(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) error {
	pref, err := uc.prefRepo.GetByAccountAndTypeAndChannel(ctx, accountID, notificationType, channel)
	if err != nil {
		return err
	}
	return uc.prefRepo.Delete(ctx, pref.ID)
}
