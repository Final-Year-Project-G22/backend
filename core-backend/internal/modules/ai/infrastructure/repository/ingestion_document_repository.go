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

type ingestionDocumentRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewIngestionDocumentRepository(db *core.Database, logger core.Logger) airepository.IngestionDocumentRepository {
	return &ingestionDocumentRepository{db: db, logger: logger}
}

func (r *ingestionDocumentRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *ingestionDocumentRepository) Create(ctx context.Context, doc *entity.IngestionDocument) error {
	if err := r.getDB(ctx).Create(doc).Error; err != nil {
		r.logger.Error("Failed to create ingestion document", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *ingestionDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.IngestionDocument, error) {
	var doc entity.IngestionDocument
	if err := r.getDB(ctx).First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFoundError("ingestion document", id)
		}
		r.logger.Error("Failed to get ingestion document by ID", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	return &doc, nil
}

func (r *ingestionDocumentRepository) GetByIdempotencyKey(ctx context.Context, accountID uuid.UUID, idempotencyKey string) (*entity.IngestionDocument, error) {
	var doc entity.IngestionDocument
	if err := r.getDB(ctx).
		Where("account_id = ? AND idempotency_key = ?", accountID, idempotencyKey).
		First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get ingestion document by idempotency key", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	return &doc, nil
}

func (r *ingestionDocumentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.IngestionDocumentStatus, lastError *string) error {
	updates := map[string]any{
		"status": status,
	}
	if lastError != nil {
		updates["last_error"] = *lastError
	}
	if err := r.getDB(ctx).Model(&entity.IngestionDocument{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to update ingestion document status", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}
	return nil
}
