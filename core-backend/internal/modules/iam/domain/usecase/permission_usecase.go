package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type PermissionUsecase interface {
	GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error)
	ListPermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)
	ListPermissions(ctx context.Context, input ListPermissionsInput) ([]*entity.Permission, error)
}

type ListPermissionsInput struct {
	Codes  []string
	Module string
}
