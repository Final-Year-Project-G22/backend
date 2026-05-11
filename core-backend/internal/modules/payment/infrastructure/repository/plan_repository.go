package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	paymentrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"gorm.io/gorm"
)

type planRepository struct {
	sharedrepo.GenericRepository[entity.Plan]
	db     *core.Database
	logger core.Logger
}

func NewPlanRepository(db *core.Database, logger core.Logger) paymentrepo.PlanRepository {
	base := sharedrepo.NewBaseRepository[entity.Plan](db, logger)
	return &planRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *planRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *planRepository) FindActiveByNameAndPeriod(ctx context.Context, name, period string) (*entity.Plan, error) {
	var plan entity.Plan
	if err := r.getDB(ctx).Where("name = ? AND period = ? AND is_active = ? AND deleted_at IS NULL", name, period, true).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) ListActive(ctx context.Context) ([]*entity.Plan, error) {
	var plans []*entity.Plan
	if err := r.getDB(ctx).Where("is_active = ? AND deleted_at IS NULL", true).Order("name, period").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}
