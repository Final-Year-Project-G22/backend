package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IngestionStatusProjectionRepository interface {
	UpsertProjection(ctx context.Context, projection *entity.IngestionStatusProjection) error
	GetByDocumentID(ctx context.Context, documentID uuid.UUID) (*entity.IngestionStatusProjection, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error)
}

type ingestionStatusProjectionRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewIngestionStatusProjectionRepository(db *core.Database, logger core.Logger) airepository.IngestionStatusProjectionRepository {
	return &ingestionStatusProjectionRepository{db: db, logger: logger}
}

func (r *ingestionStatusProjectionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *ingestionStatusProjectionRepository) UpsertProjection(ctx context.Context, projection *entity.IngestionStatusProjection) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.IngestionStatusProjection
		err := tx.Where("document_id = ?", projection.DocumentID).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if createErr := tx.Create(projection).Error; createErr != nil {
					r.logger.Error("Failed to create status projection", core.Error(createErr))
					return apperrors.InternalError("errors.databaseError", createErr)
				}
				return nil
			}
			r.logger.Error("Failed to query status projection", core.Error(err))
			return apperrors.InternalError("errors.databaseError", err)
		}

		if projection.LastEventSequence > existing.LastEventSequence {
			existing.CurrentStage = projection.CurrentStage
			existing.IsTerminal = projection.IsTerminal
			existing.UpdatedAt = projection.UpdatedAt
			existing.LastError = projection.LastError
			existing.ChunksProcessedCount = projection.ChunksProcessedCount
			existing.ChunksFailedCount = projection.ChunksFailedCount
			existing.LastEventSequence = projection.LastEventSequence
			existing.EventID = projection.EventID
			if projection.CompletedAt != nil {
				existing.CompletedAt = projection.CompletedAt
			}
			if saveErr := tx.Save(&existing).Error; saveErr != nil {
				r.logger.Error("Failed to update status projection", core.Error(saveErr))
				return apperrors.InternalError("errors.databaseError", saveErr)
			}
		}
		return nil
	})
}

func (r *ingestionStatusProjectionRepository) GetByDocumentID(ctx context.Context, documentID uuid.UUID) (*entity.IngestionStatusProjection, error) {
	var projection entity.IngestionStatusProjection
	err := r.getDB(ctx).Where("document_id = ?", documentID).First(&projection).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get status projection by document ID", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	return &projection, nil
}

func (r *ingestionStatusProjectionRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error) {
	var projections []*entity.IngestionStatusProjection
	err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projections).Error
	if err != nil {
		r.logger.Error("Failed to list status projections by account ID", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	return projections, nil
}

func (r *ingestionStatusProjectionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error) {
	var projections []*entity.IngestionStatusProjection
	err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projections).Error
	if err != nil {
		r.logger.Error("Failed to list status projections by user ID", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	return projections, nil
}
