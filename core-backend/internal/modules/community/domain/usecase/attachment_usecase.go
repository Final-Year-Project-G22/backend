package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/google/uuid"
)

type AttachmentUsecase interface {
	// Upload validates and stores multiple attachments, creating pending attachment records
	Upload(ctx context.Context, accountID uuid.UUID, inputs []AttachmentUploadInput) ([]*entity.Attachment, error)

	// LinkToPost links pre-uploaded attachments to a post after validating ownership and pending status
	LinkToPost(ctx context.Context, postID uuid.UUID, attachmentIDs []uuid.UUID, accountID uuid.UUID) error

	// UnlinkFromPost removes specific attachments from a post, deleting DB rows and storage files
	UnlinkFromPost(ctx context.Context, postID uuid.UUID, attachmentIDs []uuid.UUID, accountID uuid.UUID) error

	// DeleteOrphan permanently deletes a pending attachment (storage file + DB row)
	DeleteOrphan(ctx context.Context, attachmentID uuid.UUID, accountID uuid.UUID) error

	// CleanupPending deletes attachments where status=pending and created_at is older than the threshold
	CleanupPending(ctx context.Context, olderThanThreshold time.Time) error

	// FindByPostID retrieves all attachments for a post
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Attachment, error)
}

type AttachmentUploadInput struct {
	FileBytes []byte
	Filename  string
}
