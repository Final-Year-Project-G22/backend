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

type oauthIdentityRepository struct {
	sharedrepo.GenericRepository[entity.OAuthIdentity]
	db     *core.Database
	logger core.Logger
}

func NewOAuthIdentityRepository(db *core.Database, logger core.Logger) repository.OAuthIdentityRepository {
	base := sharedrepo.NewBaseRepository[entity.OAuthIdentity](db, logger)
	return &oauthIdentityRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *oauthIdentityRepository) GetByProviderSubject(ctx context.Context, provider, subject string) (*entity.OAuthIdentity, error) {
	var identity entity.OAuthIdentity
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_subject = ?", provider, subject).
		First(&identity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrOAuthIdentityNotFound
		}
		r.logger.Error("Failed to get OAuth identity by provider subject", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &identity, nil
}

func (r *oauthIdentityRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.OAuthIdentity, error) {
	var identities []*entity.OAuthIdentity
	if err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Find(&identities).Error; err != nil {
		r.logger.Error("Failed to list OAuth identities by account ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return identities, nil
}
