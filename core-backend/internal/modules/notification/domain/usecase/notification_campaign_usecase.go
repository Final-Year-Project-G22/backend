package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationCampaignUsecase interface {
	CreateCampaign(ctx context.Context, createdBy uuid.UUID, input CreateCampaignInput) (*entity.NotificationCampaign, error)
	GetCampaign(ctx context.Context, id uuid.UUID) (*CampaignDetail, error)
	ListCampaigns(ctx context.Context, status *entity.CampaignStatus, q query.QueryOptions) ([]*entity.NotificationCampaign, int64, error)
	UpdateCampaign(ctx context.Context, id uuid.UUID, input UpdateCampaignInput) (*entity.NotificationCampaign, error)
	ScheduleCampaign(ctx context.Context, input ScheduleCampaignInput) error
	CancelCampaign(ctx context.Context, id uuid.UUID) error
	ProcessScheduledCampaigns(ctx context.Context) error
}
