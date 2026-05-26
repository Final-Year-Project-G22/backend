package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type RolePermissionSeeder struct {
	permissionRepo     repository.PermissionRepository
	roleRepo           repository.RoleRepository
	rolePermissionRepo repository.RolePermissionRepository
	logger             core.Logger
}

func NewRolePermissionSeeder(
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	rolePermissionRepo repository.RolePermissionRepository,
	logger core.Logger,
) *RolePermissionSeeder {
	return &RolePermissionSeeder{
		permissionRepo:     permissionRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		logger:             logger,
	}
}

func (s *RolePermissionSeeder) Seed(ctx context.Context, permGroups [][]permissions.SeedPermission, roleGroups [][]permissions.SeedRole) error {
	allPermissions := flattenPermissions(permGroups)
	if err := s.seedPermissions(ctx, allPermissions); err != nil {
		return err
	}

	allRoles := flattenRoles(roleGroups)
	if err := s.seedRoles(ctx, allRoles); err != nil {
		return err
	}

	permissionCodes := make([]string, 0, len(allPermissions))
	for _, perm := range allPermissions {
		permissionCodes = append(permissionCodes, perm.Code)
	}

	permissionsByCode, err := s.loadPermissionsByCode(ctx, permissionCodes)
	if err != nil {
		return err
	}

	for _, role := range allRoles {
		seededRole, err := s.roleRepo.GetByCode(ctx, role.Code)
		if err != nil {
			return err
		}

		var targetPerms map[string]*entity.Permission
		if role.Code == "super_admin" {
			targetPerms = permissionsByCode
		} else {
			targetPerms = filterPermissionsByCodes(permissionsByCode, role.PermissionCodes)
		}

		if err := s.assignPermissions(ctx, seededRole.ID, targetPerms); err != nil {
			return err
		}
	}

	return nil
}

func (s *RolePermissionSeeder) seedPermissions(ctx context.Context, perms []permissions.SeedPermission) error {
	for _, perm := range perms {
		_, err := s.permissionRepo.GetByCode(ctx, perm.Code)
		if err == nil {
			continue
		}
		if err != iamerror.ErrPermissionNotFound {
			return err
		}

		permission := &entity.Permission{
			Code:        strings.TrimSpace(perm.Code),
			Name:        strings.TrimSpace(perm.Name),
			Description: perm.Description,
			Module:      strings.TrimSpace(perm.Module),
		}

		if err := s.permissionRepo.Create(ctx, permission); err != nil {
			return err
		}

		s.logger.Info("Permission seeded", core.String("code", permission.Code))
	}

	return nil
}

func (s *RolePermissionSeeder) seedRoles(ctx context.Context, roles []permissions.SeedRole) error {
	for _, role := range roles {
		_, err := s.roleRepo.GetByCode(ctx, role.Code)
		if err == nil {
			continue
		}
		if err != iamerror.ErrRoleNotFound {
			return err
		}

		newRole := &entity.Role{
			Code:        strings.TrimSpace(role.Code),
			Name:        strings.TrimSpace(role.Name),
			Description: role.Description,
			Type:        entity.RoleTypeSystem,
			IsSystem:    true,
			IsMutable:   false,
		}

		if err := s.roleRepo.Create(ctx, newRole); err != nil {
			return err
		}

		s.logger.Info("Role seeded", core.String("code", newRole.Code))
	}

	return nil
}

func (s *RolePermissionSeeder) loadPermissionsByCode(ctx context.Context, codes []string) (map[string]*entity.Permission, error) {
	permissions, err := s.permissionRepo.ListByCodes(ctx, codes)
	if err != nil {
		return nil, err
	}

	permissionsByCode := make(map[string]*entity.Permission, len(permissions))
	for _, permission := range permissions {
		permissionsByCode[permission.Code] = permission
	}

	if len(permissionsByCode) != len(codes) {
		missing := missingCodes(codes, permissionsByCode)
		return nil, errors.InternalError("errors.databaseError", fmt.Errorf("missing permissions: %s", strings.Join(missing, ", ")))
	}

	return permissionsByCode, nil
}

func (s *RolePermissionSeeder) assignPermissions(ctx context.Context, roleID uuid.UUID, permsByCode map[string]*entity.Permission) error {
	existing, err := s.rolePermissionRepo.ListByRoleID(ctx, roleID)
	if err != nil {
		return err
	}

	existingMap := make(map[uuid.UUID]struct{}, len(existing))
	for _, rp := range existing {
		existingMap[rp.PermissionID] = struct{}{}
	}

	toCreate := make([]*entity.RolePermission, 0, len(permsByCode))
	for _, permission := range permsByCode {
		if _, ok := existingMap[permission.ID]; ok {
			continue
		}
		toCreate = append(toCreate, &entity.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		})
	}

	if err := s.rolePermissionRepo.CreateBulk(ctx, toCreate); err != nil {
		return err
	}

	return nil
}

func flattenPermissions(groups [][]permissions.SeedPermission) []permissions.SeedPermission {
	total := 0
	for _, g := range groups {
		total += len(g)
	}
	result := make([]permissions.SeedPermission, 0, total)
	for _, g := range groups {
		result = append(result, g...)
	}
	return result
}

func flattenRoles(groups [][]permissions.SeedRole) []permissions.SeedRole {
	total := 0
	for _, g := range groups {
		total += len(g)
	}
	result := make([]permissions.SeedRole, 0, total)
	for _, g := range groups {
		result = append(result, g...)
	}
	return result
}

func filterPermissionsByCodes(permsByCode map[string]*entity.Permission, codes []string) map[string]*entity.Permission {
	result := make(map[string]*entity.Permission, len(codes))
	for _, code := range codes {
		if perm, ok := permsByCode[code]; ok {
			result[code] = perm
		}
	}
	return result
}

func missingCodes(codes []string, existing map[string]*entity.Permission) []string {
	missing := make([]string, 0)
	for _, code := range codes {
		if _, ok := existing[code]; !ok {
			missing = append(missing, code)
		}
	}
	sort.Strings(missing)
	return missing
}
