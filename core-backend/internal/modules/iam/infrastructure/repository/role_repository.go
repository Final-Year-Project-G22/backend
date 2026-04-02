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
	"gorm.io/gorm"
)

type roleRepository struct {
	sharedrepo.GenericRepository[entity.Role]
	db     *core.Database
	logger core.Logger
}

// NewRoleRepository creates a new RoleRepository implementation.
func NewRoleRepository(db *core.Database, logger core.Logger) repository.RoleRepository {
	base := sharedrepo.NewBaseRepository[entity.Role](db, logger)
	return &roleRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

// getDB returns the appropriate *gorm.DB for the context (tx-aware).
func (r *roleRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	normalizedCode := strings.TrimSpace(code)

	err := r.getDB(ctx).
		Where("code = ?", normalizedCode).
		First(&role).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrRoleNotFound
		}
		r.logger.Error("Failed to get role by code", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &role, nil
}

func (r *roleRepository) ListByCodes(ctx context.Context, codes []string) ([]*entity.Role, error) {
	if len(codes) == 0 {
		return []*entity.Role{}, nil
	}

	var roles []*entity.Role
	err := r.getDB(ctx).
		Where("code IN ?", codes).
		Find(&roles).Error

	if err != nil {
		r.logger.Error("Failed to list roles by codes", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return roles, nil
}
