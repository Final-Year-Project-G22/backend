package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationTemplateRepository interface {
	sharedrepo.GenericRepository[entity.NotificationTemplate]

	GetByType(ctx context.Context, notificationType entity.NotificationType) (*entity.NotificationTemplate, error)
	ListByTemplateGroup(ctx context.Context, group string, q query.QueryOptions) ([]*entity.NotificationTemplate, error)
	GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.NotificationTemplateTranslation, error)
	UpsertTranslation(ctx context.Context, translation *entity.NotificationTemplateTranslation) error
	DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error
}
