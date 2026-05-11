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

type paymentRepository struct {
	sharedrepo.GenericRepository[entity.Payment]
	db     *core.Database
	logger core.Logger
}

func NewPaymentRepository(db *core.Database, logger core.Logger) paymentrepo.PaymentRepository {
	base := sharedrepo.NewBaseRepository[entity.Payment](db, logger)
	return &paymentRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *paymentRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *paymentRepository) FindByTxRef(ctx context.Context, txRef string) (*entity.Payment, error) {
	var payment entity.Payment
	if err := r.getDB(ctx).Where("tx_ref = ? AND deleted_at IS NULL", txRef).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindPendingByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.Payment, error) {
	var payments []*entity.Payment
	if err := r.getDB(ctx).Where("account_id = ? AND status = ? AND deleted_at IS NULL", accountID, entity.PaymentStatusPending).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
