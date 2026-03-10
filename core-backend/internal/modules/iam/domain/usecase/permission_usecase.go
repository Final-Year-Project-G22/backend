package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type PermissionUsecase interface {
	CreatePermission(ctx context.Context, input CreatePermissionInput) (*entity.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error)
	ListPermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)
}

type CreatePermissionInput struct {
	Code        string
	Name        string
	Description *string
	Module      string
}
