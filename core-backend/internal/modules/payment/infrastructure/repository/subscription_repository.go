package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	paymentrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionRepository struct {
	sharedrepo.GenericRepository[entity.Subscription]
	db     *core.Database
	logger core.Logger
}

func NewSubscriptionRepository(db *core.Database, logger core.Logger) paymentrepo.SubscriptionRepository {
	base := sharedrepo.NewBaseRepository[entity.Subscription](db, logger)
	return &subscriptionRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *subscriptionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *subscriptionRepository) GetActiveByAccount(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error) {
	var sub entity.Subscription
	if err := r.getDB(ctx).Where("account_id = ? AND status = ? AND deleted_at IS NULL", accountID, entity.SubscriptionStatusActive).First(&sub).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetLatestByAccount(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error) {
	var sub entity.Subscription
	if err := r.getDB(ctx).Where("account_id = ? AND deleted_at IS NULL", accountID).Order("created_at DESC").First(&sub).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}
