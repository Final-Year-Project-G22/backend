package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type attachmentRepository struct {
	sharedrepo.GenericRepository[entity.Attachment]
	db     *core.Database
	logger core.Logger
}

func NewAttachmentRepository(db *core.Database, logger core.Logger) communityrepo.AttachmentRepository {
	base := sharedrepo.NewBaseRepository[entity.Attachment](db, logger)
	return &attachmentRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *attachmentRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *attachmentRepository) FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	if err := r.getDB(ctx).Where("post_id = ?", postID).Find(&attachments).Error; err != nil {
		r.logger.Error("Failed to find attachments by post ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return attachments, nil
}

func (r *attachmentRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
	if len(ids) == 0 {
		return []*entity.Attachment{}, nil
	}
	var attachments []*entity.Attachment
	if err := r.getDB(ctx).Where("id IN ?", ids).Find(&attachments).Error; err != nil {
		r.logger.Error("Failed to find attachments by IDs", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return attachments, nil
}

func (r *attachmentRepository) UpdatePostID(ctx context.Context, attachmentIDs []uuid.UUID, postID uuid.UUID) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	result := r.getDB(ctx).Model(&entity.Attachment{}).
		Where("id IN ?", attachmentIDs).
		Updates(map[string]interface{}{
			"post_id": postID,
			"status":  entity.AttachmentStatusLinked,
		})
	if result.Error != nil {
		r.logger.Error("Failed to update attachment post_id", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *attachmentRepository) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	result := r.getDB(ctx).Where("id IN ?", ids).Unscoped().Delete(&entity.Attachment{})
	if result.Error != nil {
		r.logger.Error("Failed to delete attachments by IDs", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *attachmentRepository) FindPendingOlderThan(ctx context.Context, olderThan time.Time) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	if err := r.getDB(ctx).
		Where("status = ? AND created_at < ?", entity.AttachmentStatusPending, olderThan).
		Find(&attachments).Error; err != nil {
		r.logger.Error("Failed to find pending attachments older than threshold", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return attachments, nil
}
