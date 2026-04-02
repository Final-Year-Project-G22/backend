package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rolePermissionRepository struct {
	db     *core.Database
	logger core.Logger
}

// NewRolePermissionRepository creates a new RolePermissionRepository implementation.
func NewRolePermissionRepository(db *core.Database, logger core.Logger) repository.RolePermissionRepository {
	return &rolePermissionRepository{
		db:     db,
		logger: logger,
	}
}

// getDB returns the appropriate *gorm.DB for the context (tx-aware).
func (r *rolePermissionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *entity.RolePermission) error {
	if err := r.getDB(ctx).Create(rolePermission).Error; err != nil {
		r.logger.Error("Failed to create role permission", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *rolePermissionRepository) CreateBulk(ctx context.Context, rolePermissions []*entity.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	if err := r.getDB(ctx).CreateInBatches(rolePermissions, 100).Error; err != nil {
		r.logger.Error("Failed to bulk create role permissions", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *rolePermissionRepository) ListByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entity.RolePermission, error) {
	var rolePermissions []*entity.RolePermission
	err := r.getDB(ctx).
		Where("role_id = ?", roleID).
		Find(&rolePermissions).Error

	if err != nil {
		r.logger.Error("Failed to list role permissions by role ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return rolePermissions, nil
}

func (r *rolePermissionRepository) DeleteByRoleAndPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	result := r.getDB(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&entity.RolePermission{})

	if result.Error != nil {
		r.logger.Error("Failed to delete role permission", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return iamerror.ErrPermissionNotFound
	}

	return nil
}

func (r *rolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	result := r.getDB(ctx).
		Where("role_id = ?", roleID).
		Delete(&entity.RolePermission{})

	if result.Error != nil {
		r.logger.Error("Failed to delete role permissions by role ID", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	return nil
}
