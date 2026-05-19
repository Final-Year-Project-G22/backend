package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type complianceEntryRepository struct {
	sharedrepo.GenericRepository[entity.ComplianceEntry]
	db     *core.Database
	logger core.Logger
}

func NewComplianceEntryRepository(db *core.Database, logger core.Logger) notifrepo.ComplianceEntryRepository {
	base := sharedrepo.NewBaseRepository[entity.ComplianceEntry](db, logger)
	return &complianceEntryRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *complianceEntryRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *complianceEntryRepository) ListByBusinessProfile(ctx context.Context, businessProfileID uuid.UUID) ([]*entity.ComplianceEntry, error) {
	var entries []*entity.ComplianceEntry
	if err := r.getDB(ctx).
		Where("business_profile_id = ?", businessProfileID).
		Order("expiry_date asc").
		Find(&entries).Error; err != nil {
		r.logger.Error("Failed to list compliance entries by business profile", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return entries, nil
}

func (r *complianceEntryRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.ComplianceEntry, error) {
	var entries []*entity.ComplianceEntry
	if err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Order("expiry_date asc").
		Find(&entries).Error; err != nil {
		r.logger.Error("Failed to list compliance entries by account", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return entries, nil
}

func (r *complianceEntryRepository) FetchExpiringSoon(ctx context.Context, now time.Time, limit int) ([]*entity.ComplianceEntry, error) {
	var entries []*entity.ComplianceEntry
	if err := r.getDB(ctx).
		Where("status = ? AND expiry_date - (reminder_days_before * INTERVAL '1 day') <= ?", entity.ComplianceEntryStatusActive, now).
		Where("last_notified_at IS NULL OR last_notified_at < expiry_date - (reminder_days_before * INTERVAL '1 day')").
		Order("expiry_date asc").
		Limit(limit).
		Find(&entries).Error; err != nil {
		r.logger.Error("Failed to fetch expiring compliance entries", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return entries, nil
}

func (r *complianceEntryRepository) CountByStatus(ctx context.Context, businessProfileID uuid.UUID, status entity.ComplianceEntryStatus) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.ComplianceEntry{}).
		Where("business_profile_id = ? AND status = ?", businessProfileID, status).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count compliance entries by status", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}
