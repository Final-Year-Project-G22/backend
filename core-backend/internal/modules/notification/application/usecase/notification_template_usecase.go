package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type notificationTemplateUsecase struct {
	repo             repository.NotificationTemplateRepository
	templateRenderer *service.TemplateRenderer
	transactor       sharedrepo.Transactor
}

func NewNotificationTemplateUsecase(
	repo repository.NotificationTemplateRepository,
	templateRenderer *service.TemplateRenderer,
	transactor sharedrepo.Transactor,
) usecase.NotificationTemplateUsecase {
	return &notificationTemplateUsecase{
		repo:             repo,
		templateRenderer: templateRenderer,
		transactor:       transactor,
	}
}

func (uc *notificationTemplateUsecase) CreateTemplate(ctx context.Context, input usecase.CreateTemplateInput) (*entity.NotificationTemplate, error) {
	existing, err := uc.repo.GetByType(ctx, input.NotificationType)
	if err != nil && err != notiferror.ErrTemplateNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, notiferror.ErrTemplateTypeConflict
	}

	var variablesSchema *datatypes.JSONMap
	if input.VariablesSchema != nil {
		vs := datatypes.JSONMap(*input.VariablesSchema)
		variablesSchema = &vs
	}

	tmpl := &entity.NotificationTemplate{
		Name:             input.Name,
		Description:      input.Description,
		NotificationType: input.NotificationType,
		TemplateGroup:    input.TemplateGroup,
		Priority:         input.Priority,
		DefaultContent:   datatypes.JSONMap(input.DefaultContent),
		VariablesSchema:  variablesSchema,
		DefaultTTL:       input.DefaultTTL,
	}

	if err := uc.repo.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (uc *notificationTemplateUsecase) GetTemplate(ctx context.Context, id uuid.UUID) (*entity.NotificationTemplate, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *notificationTemplateUsecase) GetTemplateByType(ctx context.Context, notificationType entity.NotificationType) (*entity.NotificationTemplate, error) {
	return uc.repo.GetByType(ctx, notificationType)
}

func (uc *notificationTemplateUsecase) ListTemplates(ctx context.Context, category *entity.NotificationCategory, q query.QueryOptions) ([]*entity.NotificationTemplate, error) {
	if category != nil {
		return uc.repo.GetByCategory(ctx, *category, q)
	}
	result := uc.repo.FindAll(ctx, q)
	return result.Data, nil
}

func (uc *notificationTemplateUsecase) UpdateTemplate(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateInput) (*entity.NotificationTemplate, error) {
	var result *entity.NotificationTemplate
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		tmpl, err := uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if tmpl.IsSystemManaged {
			return notiferror.ErrTemplateSystemManaged
		}

		updates := make(map[string]interface{})
		if input.Name != nil {
			updates["name"] = *input.Name
		}
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if input.Priority != nil {
			updates["priority"] = *input.Priority
		}
		if input.DefaultContent != nil {
			updates["default_content"] = *input.DefaultContent
		}
		if input.VariablesSchema != nil {
			updates["variables_schema"] = *input.VariablesSchema
		}
		if input.DefaultTTL != nil {
			updates["default_ttl"] = *input.DefaultTTL
		}

		if len(updates) == 0 {
			result = tmpl
			return nil
		}

		if err := uc.repo.UpdateByID(txCtx, id, updates); err != nil {
			return err
		}

		result, err = uc.repo.GetByID(txCtx, id)
		return err
	})
	return result, err
}

func (uc *notificationTemplateUsecase) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		tmpl, err := uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if tmpl.IsSystemManaged {
			return notiferror.ErrTemplateSystemManaged
		}
		return uc.repo.Delete(txCtx, id)
	})
}

func (uc *notificationTemplateUsecase) AddTranslation(ctx context.Context, input usecase.CreateTemplateTranslationInput) (*entity.NotificationTemplateTranslation, error) {
	translation := &entity.NotificationTemplateTranslation{
		TemplateID: input.TemplateID,
		Language:   input.Language,
		Subject:    input.Subject,
		Content:    input.Content,
	}
	if err := uc.repo.UpsertTranslation(ctx, translation); err != nil {
		return nil, err
	}
	return translation, nil
}

func (uc *notificationTemplateUsecase) UpdateTranslation(ctx context.Context, templateID uuid.UUID, language string, input usecase.UpdateTemplateTranslationInput) (*entity.NotificationTemplateTranslation, error) {
	translations, err := uc.repo.GetTranslations(ctx, templateID)
	if err != nil {
		return nil, err
	}

	var existing *entity.NotificationTemplateTranslation
	for _, t := range translations {
		if t.Language == language {
			existing = t
			break
		}
	}
	if existing == nil {
		return nil, notiferror.ErrTranslationNotFound
	}

	if input.Subject != nil {
		existing.Subject = *input.Subject
	}
	if input.Content != nil {
		existing.Content = *input.Content
	}

	if err := uc.repo.UpsertTranslation(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (uc *notificationTemplateUsecase) DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error {
	return uc.repo.DeleteTranslation(ctx, templateID, language)
}

func (uc *notificationTemplateUsecase) GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.NotificationTemplateTranslation, error) {
	return uc.repo.GetTranslations(ctx, templateID)
}
