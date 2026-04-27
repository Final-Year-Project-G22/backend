package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationTemplateUsecase interface {
	CreateTemplate(ctx context.Context, input CreateTemplateInput) (*entity.NotificationTemplate, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*entity.NotificationTemplate, error)
	GetTemplateByType(ctx context.Context, notificationType entity.NotificationType) (*entity.NotificationTemplate, error)
	ListTemplates(ctx context.Context, category *entity.NotificationCategory, q query.QueryOptions) ([]*entity.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, id uuid.UUID, input UpdateTemplateInput) (*entity.NotificationTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	AddTranslation(ctx context.Context, input CreateTemplateTranslationInput) (*entity.NotificationTemplateTranslation, error)
	UpdateTranslation(ctx context.Context, templateID uuid.UUID, language string, input UpdateTemplateTranslationInput) (*entity.NotificationTemplateTranslation, error)
	DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error
	GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.NotificationTemplateTranslation, error)
}
