package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type AttachmentRepository interface {
	sharedrepo.GenericRepository[entity.Attachment]

	// FindByPostID retrieves all attachments for a post
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Attachment, error)

	// FindByIDs retrieves attachments by multiple IDs
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error)

	// UpdatePostID sets the post_id on multiple attachments (links them to a post)
	UpdatePostID(ctx context.Context, attachmentIDs []uuid.UUID, postID uuid.UUID) error

	// DeleteByIDs permanently deletes multiple attachments by IDs
	DeleteByIDs(ctx context.Context, ids []uuid.UUID) error

	// FindPendingOlderThan retrieves pending attachments older than the given duration
	FindPendingOlderThan(ctx context.Context, olderThan time.Time) ([]*entity.Attachment, error)
}
