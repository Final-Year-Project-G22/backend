package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type RoleUsecase interface {
	CreateRole(ctx context.Context, input CreateRoleInput) (*entity.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*entity.Role, error)
	GetRoleByCode(ctx context.Context, code string) (*entity.Role, error)
	ListRoles(ctx context.Context) ([]*entity.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, input UpdateRoleInput) (*entity.Role, error)
	AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionCodes []string) error
	ReplaceRolePermissions(ctx context.Context, roleID uuid.UUID, permissionCodes []string) error
}

type CreateRoleInput struct {
	Code          string
	Name          string
	Description   *string
	Type          entity.RoleType
	IsSystem      bool
	IsMutable     bool
	PermissionIDs []uuid.UUID
}

type UpdateRoleInput struct {
	Name          *string
	Description   *string
	Type          *entity.RoleType
	IsSystem      *bool
	IsMutable     *bool
	PermissionIDs *[]uuid.UUID
}

type AssignRoleInput struct {
	AccountID  uuid.UUID
	RoleID     uuid.UUID
	AssignedBy uuid.UUID
	ExpiresAt  *time.Time
}
