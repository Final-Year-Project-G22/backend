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

func (s *RolePermissionSeeder) Seed(ctx context.Context) error {
	permissions := seedPermissions()
	if err := s.seedPermissions(ctx, permissions); err != nil {
		return err
	}

	roles := seedRoles()
	if err := s.seedRoles(ctx, roles); err != nil {
		return err
	}

	permissionCodes := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permissionCodes = append(permissionCodes, perm.Code)
	}

	permissionsByCode, err := s.loadPermissionsByCode(ctx, permissionCodes)
	if err != nil {
		return err
	}

	for _, role := range roles {
		seededRole, err := s.roleRepo.GetByCode(ctx, role.Code)
		if err != nil {
			return err
		}

		if role.Code == "super_admin" {
			if err := s.assignPermissions(ctx, seededRole.ID, permissionsByCode); err != nil {
				return err
			}
		} else if role.Code == "iam_admin" {
			iamAdminPerms := make(map[string]*entity.Permission)
			for code, perm := range permissionsByCode {
				if code == "iam.admin.reset_password" || code == "iam.admin.status.update" || code == "iam.role.delete" {
					continue
				}
				iamAdminPerms[code] = perm
			}
			if err := s.assignPermissions(ctx, seededRole.ID, iamAdminPerms); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *RolePermissionSeeder) seedPermissions(ctx context.Context, permissions []seedPermission) error {
	for _, perm := range permissions {
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

func (s *RolePermissionSeeder) seedRoles(ctx context.Context, roles []seedRole) error {
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

func (s *RolePermissionSeeder) assignPermissions(ctx context.Context, roleID uuid.UUID, permissionsByCode map[string]*entity.Permission) error {
	existing, err := s.rolePermissionRepo.ListByRoleID(ctx, roleID)
	if err != nil {
		return err
	}

	existingMap := make(map[uuid.UUID]struct{}, len(existing))
	for _, rolePermission := range existing {
		existingMap[rolePermission.PermissionID] = struct{}{}
	}

	toCreate := make([]*entity.RolePermission, 0, len(permissionsByCode))
	for _, permission := range permissionsByCode {
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
