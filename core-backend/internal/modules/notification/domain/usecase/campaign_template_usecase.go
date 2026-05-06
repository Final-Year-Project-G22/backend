package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CampaignTemplateUsecase interface {
	Create(ctx context.Context, input CreateCampaignTemplateInput) (*entity.CampaignTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CampaignTemplate, error)
	List(ctx context.Context, q query.QueryOptions) ([]*entity.CampaignTemplate, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateCampaignTemplateInput) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddTranslation(ctx context.Context, input CreateCampaignTemplateTranslationInput) (*entity.CampaignTemplateTranslation, error)
	UpdateTranslation(ctx context.Context, templateID uuid.UUID, language string, input UpdateCampaignTemplateTranslationInput) (*entity.CampaignTemplateTranslation, error)
	DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error
	GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.CampaignTemplateTranslation, error)
}

type CreateCampaignTemplateInput struct {
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	DefaultContent map[string]interface{} `json:"defaultContent"`
}

type UpdateCampaignTemplateInput struct {
	Name           *string                 `json:"name,omitempty"`
	Description    *string                 `json:"description,omitempty"`
	DefaultContent *map[string]interface{} `json:"defaultContent,omitempty"`
}

type CreateCampaignTemplateTranslationInput struct {
	TemplateID uuid.UUID              `json:"templateId"`
	Language   string                 `json:"language"`
	Content    map[string]interface{} `json:"content"`
}

type UpdateCampaignTemplateTranslationInput struct {
	Content *map[string]interface{} `json:"content,omitempty"`
}
