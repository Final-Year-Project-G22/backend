package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type businessProfileRepository struct {
	sharedrepo.GenericRepository[entity.BusinessProfile]
	db     *core.Database
	logger core.Logger
}

// NewBusinessProfileRepository creates a new BusinessProfileRepository implementation.
func NewBusinessProfileRepository(db *core.Database, logger core.Logger) repository.BusinessProfileRepository {
	base := sharedrepo.NewBaseRepository[entity.BusinessProfile](db, logger)
	return &businessProfileRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

// GetByAccountID retrieves a business profile by account ID.
func (r *businessProfileRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.BusinessProfile, error) {
	var profile entity.BusinessProfile
	if err := r.db.WithContext(ctx).
		Preload("Sector", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, slug, parent_id")
		}).
		Preload("Tags").
		Where("account_id = ?", accountID).
		First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrBusinessProfileNotFound
		}
		r.logger.Error("Failed to get business profile by account ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &profile, nil
}

// ExistsByAccountID checks if a business profile exists for the given account.
func (r *businessProfileRepository) ExistsByAccountID(ctx context.Context, accountID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.BusinessProfile{}).Where("account_id = ?", accountID).Count(&count).Error; err != nil {
		r.logger.Error("Failed to check business profile existence", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return count > 0, nil
}
