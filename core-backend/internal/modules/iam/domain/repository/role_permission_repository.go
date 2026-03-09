package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type RolePermissionRepository interface {
	Create(ctx context.Context, rolePermission *entity.RolePermission) error
	CreateBulk(ctx context.Context, rolePermissions []*entity.RolePermission) error
	ListByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entity.RolePermission, error)
	DeleteByRoleAndPermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error
}
