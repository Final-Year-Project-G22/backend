package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationCampaignRepository interface {
	sharedrepo.GenericRepository[entity.NotificationCampaign]

	ListByStatus(ctx context.Context, status entity.CampaignStatus, q query.QueryOptions) ([]*entity.NotificationCampaign, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.CampaignStatus) error
	ListScheduled(ctx context.Context) ([]*entity.NotificationCampaign, error)
	GetByCreator(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.NotificationCampaign, error)
}
