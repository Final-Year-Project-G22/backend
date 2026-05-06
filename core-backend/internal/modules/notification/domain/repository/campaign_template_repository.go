package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type CampaignTemplateRepository interface {
	sharedrepo.GenericRepository[entity.CampaignTemplate]

	GetTranslation(ctx context.Context, templateID uuid.UUID, language string) (*entity.CampaignTemplateTranslation, error)
	GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.CampaignTemplateTranslation, error)
	UpsertTranslation(ctx context.Context, translation *entity.CampaignTemplateTranslation) error
	DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error
}
