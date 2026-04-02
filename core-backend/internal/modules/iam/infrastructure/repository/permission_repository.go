package repository

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type permissionRepository struct {
	sharedrepo.GenericRepository[entity.Permission]
	db     *core.Database
	logger core.Logger
}

// NewPermissionRepository creates a new PermissionRepository implementation.
func NewPermissionRepository(db *core.Database, logger core.Logger) repository.PermissionRepository {
	base := sharedrepo.NewBaseRepository[entity.Permission](db, logger)
	return &permissionRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

// getDB returns the appropriate *gorm.DB for the context (tx-aware).
func (r *permissionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	var permission entity.Permission
	normalizedCode := strings.TrimSpace(code)

	err := r.getDB(ctx).
		Where("code = ?", normalizedCode).
		First(&permission).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrPermissionNotFound
		}
		r.logger.Error("Failed to get permission by code", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &permission, nil
}

func (r *permissionRepository) ListByCodes(ctx context.Context, codes []string) ([]*entity.Permission, error) {
	if len(codes) == 0 {
		return []*entity.Permission{}, nil
	}

	var permissions []*entity.Permission
	err := r.getDB(ctx).
		Where("code IN ?", codes).
		Find(&permissions).Error

	if err != nil {
		r.logger.Error("Failed to list permissions by codes", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return permissions, nil
}

func (r *permissionRepository) ListByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	var permissions []*entity.Permission

	err := r.getDB(ctx).
		Table("permissions").
		Select("permissions.*").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error

	if err != nil {
		r.logger.Error("Failed to list permissions by role ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return permissions, nil
}
