package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type NotificationIngestUsecase interface {
	ProcessEvent(ctx context.Context, input ProcessEventInput) error
	SendNotification(ctx context.Context, input SendNotificationInput) error
	SendMultiChannel(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, variables map[string]string, metadata map[string]interface{}, channels []entity.Channel, expiresAt *time.Time) error
}
