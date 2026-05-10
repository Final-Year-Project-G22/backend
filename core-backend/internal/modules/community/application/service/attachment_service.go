package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
)

// AttachmentValidator defines the interface for validating attachments
type AttachmentValidator interface {
	Validate(fileBytes []byte) (*ValidatedCommunityAttachment, error)
}

type attachmentService struct {
	attachmentRepo communityrepo.AttachmentRepository
	storage        storage.Storage
	validator      AttachmentValidator
}

func NewAttachmentService(
	attachmentRepo communityrepo.AttachmentRepository,
	storage storage.Storage,
	validator AttachmentValidator,
) usecase.AttachmentUsecase {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		storage:        storage,
		validator:      validator,
	}
}

func (s *attachmentService) Upload(ctx context.Context, accountID uuid.UUID, inputs []usecase.AttachmentUploadInput) ([]*entity.Attachment, error) {
	if len(inputs) == 0 {
		return []*entity.Attachment{}, nil
	}

	attachments := make([]*entity.Attachment, 0, len(inputs))
	for _, input := range inputs {
		validated, err := s.validator.Validate(input.FileBytes)
		if err != nil {
			s.cleanupAttachments(ctx, attachments)
			return nil, err
		}

		key := fmt.Sprintf("community/attachments/%s%s", uuid.NewString(), validated.Extension)
		uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
			Key:         key,
			Content:     validated.Content,
			ContentType: validated.ContentType,
		})
		if err != nil {
			s.cleanupAttachments(ctx, attachments)
			return nil, apperrors.InternalError("community.errors.uploadFailed", err)
		}

		url := uploaded.URL
		if url == "" {
			url = fmt.Sprintf("/api/v1/files/%s", key)
		}

		attachment := &entity.Attachment{
			StorageKey: key,
			FileURL:    url,
			FileType:   validated.ContentType,
			FileName:   input.Filename,
			UploadedBy: accountID,
			Status:     entity.AttachmentStatusPending,
		}
		if validated.Content != nil {
			size := int64(len(validated.Content))
			attachment.FileSize = &size
		}

		if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
			_ = s.storage.Delete(ctx, key)
			s.cleanupAttachments(ctx, attachments)
			return nil, err
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (s *attachmentService) LinkToPost(ctx context.Context, postID uuid.UUID, attachmentIDs []uuid.UUID, accountID uuid.UUID) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	attachments, err := s.attachmentRepo.FindByIDs(ctx, attachmentIDs)
	if err != nil {
		return err
	}

	if len(attachments) != len(attachmentIDs) {
		return apperrors.InvalidInputError("attachmentIds", "community.errors.invalidAttachment")
	}

	for _, att := range attachments {
		if att.Status != entity.AttachmentStatusPending {
			return apperrors.InvalidInputError("attachmentIds", "community.errors.attachmentNotPending")
		}
		if att.UploadedBy != accountID {
			return apperrors.ForbiddenError("community.errors.attachmentNotOwned")
		}
	}

	return s.attachmentRepo.UpdatePostID(ctx, attachmentIDs, postID)
}

func (s *attachmentService) UnlinkFromPost(ctx context.Context, postID uuid.UUID, attachmentIDs []uuid.UUID, accountID uuid.UUID) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	attachments, err := s.attachmentRepo.FindByIDs(ctx, attachmentIDs)
	if err != nil {
		return err
	}

	for _, att := range attachments {
		if att.PostID == nil || *att.PostID != postID {
			return apperrors.InvalidInputError("attachmentIds", "community.errors.attachmentNotLinked")
		}
		if att.UploadedBy != accountID {
			return apperrors.ForbiddenError("community.errors.attachmentNotOwned")
		}
	}

	for _, att := range attachments {
		_ = s.storage.Delete(ctx, att.StorageKey)
	}

	return s.attachmentRepo.DeleteByIDs(ctx, attachmentIDs)
}

func (s *attachmentService) DeleteOrphan(ctx context.Context, attachmentID uuid.UUID, accountID uuid.UUID) error {
	attachments, err := s.attachmentRepo.FindByIDs(ctx, []uuid.UUID{attachmentID})
	if err != nil {
		return err
	}
	if len(attachments) == 0 {
		return apperrors.NotFoundError("attachment", attachmentID)
	}

	att := attachments[0]
	if att.Status != entity.AttachmentStatusPending {
		return apperrors.InvalidInputError("attachmentId", "community.errors.attachmentNotPending")
	}
	if att.UploadedBy != accountID {
		return apperrors.ForbiddenError("community.errors.attachmentNotOwned")
	}

	_ = s.storage.Delete(ctx, att.StorageKey)
	return s.attachmentRepo.DeleteByIDs(ctx, []uuid.UUID{attachmentID})
}

func (s *attachmentService) CleanupPending(ctx context.Context, olderThanThreshold time.Time) error {
	attachments, err := s.attachmentRepo.FindPendingOlderThan(ctx, olderThanThreshold)
	if err != nil {
		return err
	}

	for _, att := range attachments {
		_ = s.storage.Delete(ctx, att.StorageKey)
	}

	ids := make([]uuid.UUID, 0, len(attachments))
	for _, att := range attachments {
		ids = append(ids, att.ID)
	}

	return s.attachmentRepo.DeleteByIDs(ctx, ids)
}

func (s *attachmentService) FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Attachment, error) {
	return s.attachmentRepo.FindByPostID(ctx, postID)
}

func (s *attachmentService) cleanupAttachments(ctx context.Context, attachments []*entity.Attachment) {
	for _, att := range attachments {
		if att.StorageKey != "" {
			_ = s.storage.Delete(ctx, att.StorageKey)
		}
		if att.ID != uuid.Nil {
			_ = s.attachmentRepo.Delete(ctx, att.ID)
		}
	}
}
