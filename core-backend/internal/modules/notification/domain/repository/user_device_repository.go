package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type UserDeviceRepository interface {
	sharedrepo.GenericRepository[entity.UserDevice]

	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.UserDevice, error)
	GetByDeviceToken(ctx context.Context, deviceToken string) (*entity.UserDevice, error)
	DeactivateByAccount(ctx context.Context, accountID uuid.UUID) error
	UpdatePushToken(ctx context.Context, id uuid.UUID, pushToken string) error
}
