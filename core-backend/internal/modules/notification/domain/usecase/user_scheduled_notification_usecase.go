package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type UserScheduledNotificationUsecase interface {
	Schedule(ctx context.Context, accountID uuid.UUID, input ScheduleUserNotificationInput) (*entity.UserScheduledNotification, error)
	List(ctx context.Context, accountID uuid.UUID) ([]*entity.UserScheduledNotification, error)
	GetByID(ctx context.Context, accountID uuid.UUID, id uuid.UUID) (*entity.UserScheduledNotification, error)
	Cancel(ctx context.Context, accountID uuid.UUID, id uuid.UUID) error
	Reschedule(ctx context.Context, accountID uuid.UUID, id uuid.UUID, input RescheduleUserNotificationInput) error
	ListTemplates(ctx context.Context) ([]*entity.ScheduledAlertTemplate, error)
}
