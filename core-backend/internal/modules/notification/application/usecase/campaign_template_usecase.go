package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type campaignTemplateUsecase struct {
	repo       repository.CampaignTemplateRepository
	transactor sharedrepo.Transactor
	logger     core.Logger
}

func NewCampaignTemplateUsecase(
	repo repository.CampaignTemplateRepository,
	transactor sharedrepo.Transactor,
	logger core.Logger,
) usecase.CampaignTemplateUsecase {
	return &campaignTemplateUsecase{
		repo:       repo,
		transactor: transactor,
		logger:     logger,
	}
}

func (uc *campaignTemplateUsecase) Create(ctx context.Context, input usecase.CreateCampaignTemplateInput) (*entity.CampaignTemplate, error) {
	tmpl := &entity.CampaignTemplate{
		Name:           input.Name,
		Description:    input.Description,
		DefaultContent: datatypes.JSONMap(input.DefaultContent),
	}
	if err := uc.repo.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (uc *campaignTemplateUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.CampaignTemplate, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *campaignTemplateUsecase) List(ctx context.Context, q query.QueryOptions) ([]*entity.CampaignTemplate, int64, error) {
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	templates, err := uc.repo.Find(ctx, q)
	return templates, total, err
}

func (uc *campaignTemplateUsecase) Update(ctx context.Context, id uuid.UUID, input usecase.UpdateCampaignTemplateInput) error {
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	if input.DefaultContent != nil {
		updates["default_content"] = datatypes.JSONMap(*input.DefaultContent)
	}

	if len(updates) == 0 {
		return nil
	}
	return uc.repo.UpdateByID(ctx, id, updates)
}

func (uc *campaignTemplateUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *campaignTemplateUsecase) AddTranslation(ctx context.Context, input usecase.CreateCampaignTemplateTranslationInput) (*entity.CampaignTemplateTranslation, error) {
	translation := &entity.CampaignTemplateTranslation{
		CampaignTemplateID: input.TemplateID,
		Language:           input.Language,
		Content:            datatypes.JSONMap(input.Content),
	}
	if err := uc.repo.UpsertTranslation(ctx, translation); err != nil {
		return nil, err
	}
	return translation, nil
}

func (uc *campaignTemplateUsecase) UpdateTranslation(ctx context.Context, templateID uuid.UUID, language string, input usecase.UpdateCampaignTemplateTranslationInput) (*entity.CampaignTemplateTranslation, error) {
	translations, err := uc.repo.GetTranslations(ctx, templateID)
	if err != nil {
		return nil, err
	}

	var existing *entity.CampaignTemplateTranslation
	for _, t := range translations {
		if t.Language == language {
			existing = t
			break
		}
	}
	if existing == nil {
		return nil, notiferror.ErrCampaignTranslationNotFound
	}

	if input.Content != nil {
		existing.Content = datatypes.JSONMap(*input.Content)
	}

	if err := uc.repo.UpsertTranslation(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (uc *campaignTemplateUsecase) DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error {
	return uc.repo.DeleteTranslation(ctx, templateID, language)
}

func (uc *campaignTemplateUsecase) GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.CampaignTemplateTranslation, error) {
	return uc.repo.GetTranslations(ctx, templateID)
}
