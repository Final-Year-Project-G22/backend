package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type NotificationPreferenceUsecase interface {
	SetPreference(ctx context.Context, accountID uuid.UUID, input SetPreferenceInput) error
	GetPreferences(ctx context.Context, accountID uuid.UUID) ([]*entity.UserNotificationPreference, error)
	GetEffectivePreference(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (bool, error)
	IsQuietHours(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (bool, error)
	DeletePreference(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) error
}
