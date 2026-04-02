package appusecase

import (
	"context"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type roleAssignmentUsecase struct {
	roleAssignmentRepo repository.RoleAssignmentRepository
	roleRepo           repository.RoleRepository
	permissionRepo     repository.PermissionRepository
	logger             core.Logger
}

func NewRoleAssignmentUsecase(
	roleAssignmentRepo repository.RoleAssignmentRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	logger core.Logger,
) usecase.RoleAssignmentUsecase {
	return &roleAssignmentUsecase{
		roleAssignmentRepo: roleAssignmentRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		logger:             logger,
	}
}

func (u *roleAssignmentUsecase) AssignRole(ctx context.Context, input usecase.AssignRoleInput) (*entity.RoleAssignment, error) {
	if _, err := u.roleRepo.GetByID(ctx, input.RoleID); err != nil {
		return nil, err
	}

	exists, err := u.roleAssignmentRepo.ExistsByAccountAndRole(ctx, input.AccountID, input.RoleID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ConflictError("iam.errors.conflict")
	}

	assignment := &entity.RoleAssignment{
		AccountID:  input.AccountID,
		RoleID:     input.RoleID,
		AssignedBy: input.AssignedBy,
		ExpiresAt:  input.ExpiresAt,
	}

	if err := u.roleAssignmentRepo.Create(ctx, assignment); err != nil {
		return nil, err
	}

	u.logger.Info("Role assigned",
		core.String("roleAssignmentID", assignment.ID.String()),
		core.String("accountID", input.AccountID.String()),
		core.String("roleID", input.RoleID.String()),
	)

	return assignment, nil
}

func (u *roleAssignmentUsecase) RevokeRole(ctx context.Context, assignmentID uuid.UUID, reason *string) error {
	err := u.roleAssignmentRepo.Revoke(ctx, assignmentID, time.Now(), reason)
	if err == iamerror.ErrRoleNotAssigned {
		return errors.NotFoundErrorWithKey("iam.errors.notFound")
	}
	if err != nil {
		return err
	}

	u.logger.Info("Role revoked", core.String("roleAssignmentID", assignmentID.String()))
	return nil
}

func (u *roleAssignmentUsecase) ListAccountRoles(ctx context.Context, accountID uuid.UUID) ([]*entity.Role, error) {
	assignments, err := u.roleAssignmentRepo.ListByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	roles := make([]*entity.Role, 0, len(assignments))
	for _, assignment := range assignments {
		if assignment.Role.ID != uuid.Nil {
			role := assignment.Role
			roles = append(roles, &role)
			continue
		}
		role, err := u.roleRepo.GetByID(ctx, assignment.RoleID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (u *roleAssignmentUsecase) ListAccountRoleAssignments(ctx context.Context, accountID uuid.UUID) ([]*entity.RoleAssignment, error) {
	return u.roleAssignmentRepo.ListByAccountID(ctx, accountID)
}

func (u *roleAssignmentUsecase) GetEffectivePermissions(ctx context.Context, accountID uuid.UUID) ([]*entity.Permission, error) {
	roles, err := u.ListAccountRoles(ctx, accountID)
	if err != nil {
		return nil, err
	}

	permissionsMap := make(map[uuid.UUID]*entity.Permission)
	for _, role := range roles {
		perms, err := u.permissionRepo.ListByRoleID(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, perm := range perms {
			permissionsMap[perm.ID] = perm
		}
	}

	permissions := make([]*entity.Permission, 0, len(permissionsMap))
	for _, perm := range permissionsMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (u *roleAssignmentUsecase) HasPermission(ctx context.Context, accountID uuid.UUID, permissionCode string) (bool, error) {
	permissions, err := u.GetEffectivePermissions(ctx, accountID)
	if err != nil {
		return false, err
	}

	normalizedCode := strings.TrimSpace(permissionCode)
	for _, perm := range permissions {
		if strings.EqualFold(perm.Code, normalizedCode) {
			return true, nil
		}
	}

	return false, nil
}
