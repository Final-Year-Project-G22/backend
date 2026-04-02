package appusecase

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type roleUsecase struct {
	roleRepo           repository.RoleRepository
	permissionRepo     repository.PermissionRepository
	rolePermissionRepo repository.RolePermissionRepository
	logger             core.Logger
}

func NewRoleUsecase(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	rolePermissionRepo repository.RolePermissionRepository,
	logger core.Logger,
) usecase.RoleUsecase {
	return &roleUsecase{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		logger:             logger,
	}
}

func (u *roleUsecase) CreateRole(ctx context.Context, input usecase.CreateRoleInput) (*entity.Role, error) {
	code := strings.TrimSpace(input.Code)
	name := strings.TrimSpace(input.Name)

	existing, err := u.roleRepo.GetByCode(ctx, code)
	if err == nil && existing != nil {
		return nil, errors.ConflictError("iam.errors.conflict")
	}
	if err != nil && err != iamerror.ErrRoleNotFound {
		return nil, err
	}

	roleType := input.Type
	if roleType == "" {
		roleType = entity.RoleTypeSystem
	}

	role := &entity.Role{
		Code:        code,
		Name:        name,
		Description: input.Description,
		Type:        roleType,
		IsSystem:    input.IsSystem,
		IsMutable:   input.IsMutable,
	}

	if err := u.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	if len(input.PermissionIDs) > 0 {
		permissions, err := u.getPermissionsByIDs(ctx, input.PermissionIDs)
		if err != nil {
			return nil, err
		}
		rolePermissions := buildRolePermissions(role.ID, permissions)
		if err := u.rolePermissionRepo.CreateBulk(ctx, rolePermissions); err != nil {
			return nil, err
		}
	}

	u.logger.Info("Role created", core.String("roleID", role.ID.String()))
	return role, nil
}

func (u *roleUsecase) GetRole(ctx context.Context, roleID uuid.UUID) (*entity.Role, error) {
	return u.roleRepo.GetByID(ctx, roleID)
}

func (u *roleUsecase) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	role, err := u.roleRepo.GetByCode(ctx, code)
	if err == iamerror.ErrRoleNotFound {
		return nil, errors.NotFoundErrorWithKey("iam.errors.roleNotFound")
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (u *roleUsecase) ListRoles(ctx context.Context) ([]*entity.Role, error) {
	return u.roleRepo.Find(ctx, query.QueryOptions{})
}

func (u *roleUsecase) UpdateRole(ctx context.Context, roleID uuid.UUID, input usecase.UpdateRoleInput) (*entity.Role, error) {
	role, err := u.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		role.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		role.Description = input.Description
	}
	if input.Type != nil {
		role.Type = *input.Type
	}
	if input.IsSystem != nil {
		role.IsSystem = *input.IsSystem
	}
	if input.IsMutable != nil {
		role.IsMutable = *input.IsMutable
	}

	if err := u.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	if input.PermissionIDs != nil {
		if err := u.rolePermissionRepo.DeleteByRoleID(ctx, roleID); err != nil {
			return nil, err
		}

		if len(*input.PermissionIDs) > 0 {
			permissions, err := u.getPermissionsByIDs(ctx, *input.PermissionIDs)
			if err != nil {
				return nil, err
			}
			rolePermissions := buildRolePermissions(role.ID, permissions)
			if err := u.rolePermissionRepo.CreateBulk(ctx, rolePermissions); err != nil {
				return nil, err
			}
		}
	}

	u.logger.Info("Role updated", core.String("roleID", role.ID.String()))
	return role, nil
}

func (u *roleUsecase) AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionCodes []string) error {
	if len(permissionCodes) == 0 {
		return nil
	}

	if _, err := u.roleRepo.GetByID(ctx, roleID); err != nil {
		return err
	}

	permissions, err := u.permissionRepo.ListByCodes(ctx, permissionCodes)
	if err != nil {
		return err
	}

	if len(permissions) != len(permissionCodes) {
		return errors.NotFoundErrorWithKey("iam.errors.notFound")
	}

	existing, err := u.rolePermissionRepo.ListByRoleID(ctx, roleID)
	if err != nil {
		return err
	}

	existingMap := make(map[uuid.UUID]struct{}, len(existing))
	for _, rp := range existing {
		existingMap[rp.PermissionID] = struct{}{}
	}

	toCreate := make([]*entity.RolePermission, 0, len(permissions))
	for _, permission := range permissions {
		if _, ok := existingMap[permission.ID]; ok {
			continue
		}
		toCreate = append(toCreate, &entity.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		})
	}

	return u.rolePermissionRepo.CreateBulk(ctx, toCreate)
}

func (u *roleUsecase) ReplaceRolePermissions(ctx context.Context, roleID uuid.UUID, permissionCodes []string) error {
	if _, err := u.roleRepo.GetByID(ctx, roleID); err != nil {
		return err
	}

	if err := u.rolePermissionRepo.DeleteByRoleID(ctx, roleID); err != nil {
		return err
	}

	if len(permissionCodes) == 0 {
		return nil
	}

	permissions, err := u.permissionRepo.ListByCodes(ctx, permissionCodes)
	if err != nil {
		return err
	}

	if len(permissions) != len(permissionCodes) {
		return errors.NotFoundErrorWithKey("iam.errors.notFound")
	}

	rolePermissions := make([]*entity.RolePermission, 0, len(permissions))
	for _, permission := range permissions {
		rolePermissions = append(rolePermissions, &entity.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		})
	}

	return u.rolePermissionRepo.CreateBulk(ctx, rolePermissions)
}

func (u *roleUsecase) getPermissionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Permission, error) {
	permissions, err := u.permissionRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	uniqueIDs := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		uniqueIDs[id] = struct{}{}
	}

	if len(permissions) != len(uniqueIDs) {
		return nil, errors.NotFoundErrorWithKey("iam.errors.notFound")
	}

	return permissions, nil
}

func buildRolePermissions(roleID uuid.UUID, permissions []*entity.Permission) []*entity.RolePermission {
	rolePermissions := make([]*entity.RolePermission, 0, len(permissions))
	seen := make(map[uuid.UUID]struct{}, len(permissions))
	for _, permission := range permissions {
		if _, ok := seen[permission.ID]; ok {
			continue
		}
		seen[permission.ID] = struct{}{}
		rolePermissions = append(rolePermissions, &entity.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		})
	}

	return rolePermissions
}
