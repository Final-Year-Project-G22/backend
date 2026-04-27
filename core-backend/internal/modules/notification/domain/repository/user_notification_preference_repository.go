package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type UserNotificationPreferenceRepository interface {
	sharedrepo.GenericRepository[entity.UserNotificationPreference]

	GetByAccountAndTypeAndChannel(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (*entity.UserNotificationPreference, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.UserNotificationPreference, error)
	Upsert(ctx context.Context, pref *entity.UserNotificationPreference) error
}
