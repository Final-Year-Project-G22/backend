package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type EmailDeliveryLogRepository interface {
	sharedrepo.GenericRepository[entity.EmailDeliveryLog]

	GetByProviderMessageID(ctx context.Context, providerMessageID string) (*entity.EmailDeliveryLog, error)
	UpdateDeliveryEvent(ctx context.Context, id uuid.UUID, eventType string, occurredAt time.Time, metadata map[string]interface{}) error
	GetByNotificationHistoryID(ctx context.Context, historyID uuid.UUID) (*entity.EmailDeliveryLog, error)
}
