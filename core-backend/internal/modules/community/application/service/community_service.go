package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
)

type AttachmentUploadInput struct {
	FileBytes []byte
	Filename  string
}

type CommunityService interface {
	CreateThreadWithPost(ctx context.Context, accountID uuid.UUID, input usecase.CreateThreadInput, attachment *AttachmentUploadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error)
	ReplyToThread(ctx context.Context, accountID, threadID uuid.UUID, parentPostID *uuid.UUID, input usecase.CreatePostInput, attachment *AttachmentUploadInput) (*entity.DiscussionPost, error)
	UpdatePost(ctx context.Context, accountID, postID uuid.UUID, input usecase.UpdatePostInput, attachment *AttachmentUploadInput) (*entity.DiscussionPost, error)
	MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error
	BlockUserInThread(ctx context.Context, input usecase.BlockUserInput) error
	UnblockUserInThread(ctx context.Context, input usecase.BlockUserInput) error
	FollowThread(ctx context.Context, accountID, threadID uuid.UUID) error
	UnfollowThread(ctx context.Context, accountID, threadID uuid.UUID) error
	FollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	UnfollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	ListFollowedThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error)
	ListFollowedCategories(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error)
	ReportThread(ctx context.Context, reporterID uuid.UUID, input usecase.ReportThreadInput) (*entity.ContentReport, error)
	ReportPost(ctx context.Context, reporterID uuid.UUID, input usecase.ReportPostInput) (*entity.ContentReport, error)
	ReportUser(ctx context.Context, reporterID uuid.UUID, input usecase.ReportUserInput) (*entity.ContentReport, error)
}

type communityService struct {
	threadUsecase usecase.DiscussionThreadUsecase
	postUsecase   usecase.DiscussionPostUsecase
	blockUsecase  usecase.ThreadBlockUsecase
	followUsecase usecase.CommunityFollowUsecase
	reportUsecase usecase.ContentReportUsecase
	storage       storage.Storage
	validator     *CommunityAttachmentValidator
}

func NewCommunityService(
	threadUsecase usecase.DiscussionThreadUsecase,
	postUsecase usecase.DiscussionPostUsecase,
	blockUsecase usecase.ThreadBlockUsecase,
	followUsecase usecase.CommunityFollowUsecase,
	reportUsecase usecase.ContentReportUsecase,
	storage storage.Storage,
	validator *CommunityAttachmentValidator,
) CommunityService {
	return &communityService{
		threadUsecase: threadUsecase,
		postUsecase:   postUsecase,
		blockUsecase:  blockUsecase,
		followUsecase: followUsecase,
		reportUsecase: reportUsecase,
		storage:       storage,
		validator:     validator,
	}
}

func (s *communityService) CreateThreadWithPost(ctx context.Context, accountID uuid.UUID, input usecase.CreateThreadInput, attachment *AttachmentUploadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error) {
	key, attachmentURL, attachmentType, err := s.uploadAttachment(ctx, attachment)
	if err != nil {
		return nil, nil, err
	}

	input.AttachmentURL = attachmentURL
	input.AttachmentType = attachmentType

	thread, post, err := s.threadUsecase.CreateThread(ctx, accountID, input)
	if err != nil {
		s.cleanupUploadedAttachment(ctx, key)
		return nil, nil, err
	}

	return thread, post, nil
}

func (s *communityService) ReplyToThread(ctx context.Context, accountID, threadID uuid.UUID, parentPostID *uuid.UUID, input usecase.CreatePostInput, attachment *AttachmentUploadInput) (*entity.DiscussionPost, error) {
	key, attachmentURL, attachmentType, err := s.uploadAttachment(ctx, attachment)
	if err != nil {
		return nil, err
	}

	input.AttachmentURL = attachmentURL
	input.AttachmentType = attachmentType

	if parentPostID != nil {
		post, err := s.postUsecase.ReplyToPost(ctx, accountID, threadID, *parentPostID, input)
		if err != nil {
			s.cleanupUploadedAttachment(ctx, key)
			return nil, err
		}
		return post, nil
	}

	post, err := s.postUsecase.CreatePost(ctx, accountID, threadID, input)
	if err != nil {
		s.cleanupUploadedAttachment(ctx, key)
		return nil, err
	}

	return post, nil
}

func (s *communityService) MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error {
	return s.postUsecase.MarkSolution(ctx, accountID, threadID, postID)
}

func (s *communityService) UpdatePost(ctx context.Context, accountID, postID uuid.UUID, input usecase.UpdatePostInput, attachment *AttachmentUploadInput) (*entity.DiscussionPost, error) {
	if attachment != nil && input.RemoveAttachment != nil && *input.RemoveAttachment {
		return nil, apperrors.InvalidInputError("removeAttachment", "community.errors.invalidInput")
	}

	current, err := s.postUsecase.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	newKey, attachmentURL, attachmentType, err := s.uploadAttachment(ctx, attachment)
	if err != nil {
		return nil, err
	}

	if attachmentURL != nil {
		input.AttachmentURL = attachmentURL
		input.AttachmentType = attachmentType
	}

	updated, err := s.postUsecase.UpdatePost(ctx, accountID, postID, input)
	if err != nil {
		s.cleanupUploadedAttachment(ctx, newKey)
		return nil, err
	}

	if current.AttachmentURL != nil {
		shouldCleanupOld := newKey != "" || (input.RemoveAttachment != nil && *input.RemoveAttachment)
		if shouldCleanupOld {
			oldKey := s.extractKeyFromURL(*current.AttachmentURL)
			if oldKey != "" && oldKey != newKey {
				s.cleanupUploadedAttachment(ctx, oldKey)
			}
		}
	}

	return updated, nil
}

func (s *communityService) BlockUserInThread(ctx context.Context, input usecase.BlockUserInput) error {
	return s.blockUsecase.BlockUser(ctx, input)
}

func (s *communityService) UnblockUserInThread(ctx context.Context, input usecase.BlockUserInput) error {
	return s.blockUsecase.UnblockUser(ctx, input)
}

func (s *communityService) FollowThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	return s.followUsecase.FollowThread(ctx, accountID, threadID)
}

func (s *communityService) UnfollowThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	return s.followUsecase.UnfollowThread(ctx, accountID, threadID)
}

func (s *communityService) FollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error {
	return s.followUsecase.FollowCategory(ctx, accountID, categoryID)
}

func (s *communityService) UnfollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error {
	return s.followUsecase.UnfollowCategory(ctx, accountID, categoryID)
}

func (s *communityService) ListFollowedThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error) {
	return s.followUsecase.ListFollowedThreads(ctx, accountID, q)
}

func (s *communityService) ListFollowedCategories(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error) {
	return s.followUsecase.ListFollowedCategories(ctx, accountID, q)
}

func (s *communityService) ReportThread(ctx context.Context, reporterID uuid.UUID, input usecase.ReportThreadInput) (*entity.ContentReport, error) {
	return s.reportUsecase.ReportThread(ctx, reporterID, input)
}

func (s *communityService) ReportPost(ctx context.Context, reporterID uuid.UUID, input usecase.ReportPostInput) (*entity.ContentReport, error) {
	return s.reportUsecase.ReportPost(ctx, reporterID, input)
}

func (s *communityService) ReportUser(ctx context.Context, reporterID uuid.UUID, input usecase.ReportUserInput) (*entity.ContentReport, error) {
	return s.reportUsecase.ReportUser(ctx, reporterID, input)
}

func (s *communityService) uploadAttachment(ctx context.Context, attachment *AttachmentUploadInput) (string, *string, *string, error) {
	if attachment == nil {
		return "", nil, nil, nil
	}

	validated, err := s.validator.Validate(attachment.FileBytes)
	if err != nil {
		return "", nil, nil, err
	}

	key := fmt.Sprintf("community/attachments/%s%s", uuid.NewString(), validated.Extension)
	uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
		Key:         key,
		Content:     validated.Content,
		ContentType: validated.ContentType,
	})
	if err != nil {
		return "", nil, nil, apperrors.InternalError("community.errors.uploadFailed", err)
	}

	url := uploaded.URL
	if url == "" {
		url = fmt.Sprintf("/api/v1/files/%s", key)
	}

	contentType := validated.ContentType
	return key, &url, &contentType, nil
}

func (s *communityService) cleanupUploadedAttachment(ctx context.Context, key string) {
	if key == "" {
		return
	}
	_ = s.storage.Delete(ctx, key)
}

func (s *communityService) extractKeyFromURL(url string) string {
	if url == "" {
		return ""
	}

	if strings.Contains(url, "/api/v1/files/") {
		parts := strings.Split(url, "/api/v1/files/")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}

	candidate := strings.Join(parts[len(parts)-2:], "/")
	if strings.HasPrefix(candidate, "community/") {
		return candidate
	}

	return ""
}
