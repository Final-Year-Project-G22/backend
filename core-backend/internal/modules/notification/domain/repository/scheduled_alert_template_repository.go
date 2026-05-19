package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type ScheduledAlertTemplateRepository interface {
	sharedrepo.GenericRepository[entity.ScheduledAlertTemplate]

	ListActive(ctx context.Context) ([]*entity.ScheduledAlertTemplate, error)
	GetBySlug(ctx context.Context, slug string) (*entity.ScheduledAlertTemplate, error)
}
