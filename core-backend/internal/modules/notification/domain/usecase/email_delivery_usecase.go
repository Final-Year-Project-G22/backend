package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type EmailDeliveryUsecase interface {
	HandleWebhookEvent(ctx context.Context, event ResendWebhookEvent) error
	GetDeliveryLog(ctx context.Context, historyID uuid.UUID) (*entity.EmailDeliveryLog, error)
	GetDeliveryLogByProviderID(ctx context.Context, providerMessageID string) (*entity.EmailDeliveryLog, error)
}
