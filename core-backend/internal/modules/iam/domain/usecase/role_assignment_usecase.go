package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type RoleAssignmentUsecase interface {
	AssignRole(ctx context.Context, input AssignRoleInput) (*entity.RoleAssignment, error)
	RevokeRole(ctx context.Context, assignmentID uuid.UUID, reason *string) error
	ListAccountRoles(ctx context.Context, accountID uuid.UUID) ([]*entity.Role, error)
	ListAccountRoleAssignments(ctx context.Context, accountID uuid.UUID) ([]*entity.RoleAssignment, error)
	GetEffectivePermissions(ctx context.Context, accountID uuid.UUID) ([]*entity.Permission, error)
	HasPermission(ctx context.Context, accountID uuid.UUID, permissionCode string) (bool, error)
}
