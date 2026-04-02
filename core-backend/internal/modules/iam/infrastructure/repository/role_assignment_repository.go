package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type roleAssignmentRepository struct {
	sharedrepo.GenericRepository[entity.RoleAssignment]
	db     *core.Database
	logger core.Logger
}

// NewRoleAssignmentRepository creates a new RoleAssignmentRepository implementation.
func NewRoleAssignmentRepository(db *core.Database, logger core.Logger) repository.RoleAssignmentRepository {
	base := sharedrepo.NewBaseRepository[entity.RoleAssignment](db, logger)
	return &roleAssignmentRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

// getDB returns the appropriate *gorm.DB for the context (tx-aware).
func (r *roleAssignmentRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *roleAssignmentRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.RoleAssignment, error) {
	var assignments []*entity.RoleAssignment

	err := r.getDB(ctx).
		Preload("Role").
		Where("account_id = ?", accountID).
		Where("revoked_at IS NULL").
		Find(&assignments).Error

	if err != nil {
		r.logger.Error("Failed to list role assignments by account ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return assignments, nil
}

func (r *roleAssignmentRepository) ExistsByAccountAndRole(ctx context.Context, accountID, roleID uuid.UUID) (bool, error) {
	var count int64
	err := r.getDB(ctx).
		Model(&entity.RoleAssignment{}).
		Where("account_id = ?", accountID).
		Where("role_id = ?", roleID).
		Where("revoked_at IS NULL").
		Count(&count).Error

	if err != nil {
		r.logger.Error("Failed to check role assignment existence", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}

	return count > 0, nil
}

func (r *roleAssignmentRepository) GetByAccountAndRole(ctx context.Context, accountID, roleID uuid.UUID) (*entity.RoleAssignment, error) {
	var assignment entity.RoleAssignment

	err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Where("role_id = ?", roleID).
		Where("revoked_at IS NULL").
		First(&assignment).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrRoleNotAssigned
		}
		r.logger.Error("Failed to get role assignment", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &assignment, nil
}

func (r *roleAssignmentRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time, reason *string) error {
	updates := map[string]interface{}{
		"revoked_at":    revokedAt,
		"revoke_reason": reason,
	}

	result := r.getDB(ctx).
		Model(&entity.RoleAssignment{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("Failed to revoke role assignment", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	if result.RowsAffected == 0 {
		return iamerror.ErrRoleNotAssigned
	}

	return nil
}
