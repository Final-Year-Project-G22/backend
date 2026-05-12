package service

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/ws"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CommunityService interface {
	CreateThreadWithPost(ctx context.Context, accountID uuid.UUID, input usecase.CreateThreadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error)
	ReplyToThread(ctx context.Context, accountID, threadID uuid.UUID, parentPostID *uuid.UUID, input usecase.CreatePostInput) (*entity.DiscussionPost, error)
	UpdatePost(ctx context.Context, accountID, postID uuid.UUID, input usecase.UpdatePostInput) (*entity.DiscussionPost, error)
	MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error
	BlockUserInThread(ctx context.Context, input usecase.BlockUserInput) error
	UnblockUserInThread(ctx context.Context, input usecase.BlockUserInput) error
	FollowThread(ctx context.Context, accountID, threadID uuid.UUID) error
	UnfollowThread(ctx context.Context, accountID, threadID uuid.UUID) error
	FollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	UnfollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	ListFollowedThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error)
	ListFollowedCategories(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error)
	ListThreadFollowStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	ListThreadUnreadCounts(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]int, error)
	ListThreadSolutionStatus(ctx context.Context, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	MarkThreadRead(ctx context.Context, accountID, threadID uuid.UUID) error
	RecordThreadView(ctx context.Context, accountID, threadID uuid.UUID) error
	NotifyThreadUpdated(ctx context.Context, threadID uuid.UUID)
	ReportThread(ctx context.Context, reporterID uuid.UUID, input usecase.ReportThreadInput) (*entity.ContentReport, error)
	ReportPost(ctx context.Context, reporterID uuid.UUID, input usecase.ReportPostInput) (*entity.ContentReport, error)
	ReportUser(ctx context.Context, reporterID uuid.UUID, input usecase.ReportUserInput) (*entity.ContentReport, error)
}

type communityService struct {
	threadUsecase     usecase.DiscussionThreadUsecase
	postUsecase       usecase.DiscussionPostUsecase
	attachmentUsecase usecase.AttachmentUsecase
	blockUsecase      usecase.ThreadBlockUsecase
	followUsecase     usecase.CommunityFollowUsecase
	reportUsecase     usecase.ContentReportUsecase
	threadRepo        repository.DiscussionThreadRepository
	wsHub             *ws.Hub
}

func NewCommunityService(
	threadUsecase usecase.DiscussionThreadUsecase,
	postUsecase usecase.DiscussionPostUsecase,
	attachmentUsecase usecase.AttachmentUsecase,
	blockUsecase usecase.ThreadBlockUsecase,
	followUsecase usecase.CommunityFollowUsecase,
	reportUsecase usecase.ContentReportUsecase,
	threadRepo repository.DiscussionThreadRepository,
	wsHub *ws.Hub,
) CommunityService {
	return &communityService{
		threadUsecase:     threadUsecase,
		postUsecase:       postUsecase,
		attachmentUsecase: attachmentUsecase,
		blockUsecase:      blockUsecase,
		followUsecase:     followUsecase,
		reportUsecase:     reportUsecase,
		threadRepo:        threadRepo,
		wsHub:             wsHub,
	}
}

func (s *communityService) CreateThreadWithPost(ctx context.Context, accountID uuid.UUID, input usecase.CreateThreadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error) {
	thread, post, err := s.threadUsecase.CreateThread(ctx, accountID, input)
	if err != nil {
		return nil, nil, err
	}

	if len(input.AttachmentIds) > 0 {
		if err := s.attachmentUsecase.LinkToPost(ctx, post.ID, input.AttachmentIds, accountID); err != nil {
			return nil, nil, err
		}
	}

	return thread, post, nil
}

func (s *communityService) ReplyToThread(ctx context.Context, accountID, threadID uuid.UUID, parentPostID *uuid.UUID, input usecase.CreatePostInput) (*entity.DiscussionPost, error) {
	var post *entity.DiscussionPost
	var err error
	if parentPostID != nil {
		post, err = s.postUsecase.ReplyToPost(ctx, accountID, threadID, *parentPostID, input)
	} else {
		post, err = s.postUsecase.CreatePost(ctx, accountID, threadID, input)
	}
	if err != nil {
		return nil, err
	}
	if len(input.AttachmentIds) > 0 {
		if err := s.attachmentUsecase.LinkToPost(ctx, post.ID, input.AttachmentIds, accountID); err != nil {
			return nil, err
		}
	}

	_ = s.followUsecase.FollowThread(ctx, accountID, threadID)

	s.wsHub.PublishToThread(threadID, ws.NewPostCreatedEvent(threadID.String(), post.ID.String()))

	return post, nil
}

func (s *communityService) MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error {
	return s.postUsecase.MarkSolution(ctx, accountID, threadID, postID)
}

func (s *communityService) UpdatePost(ctx context.Context, accountID, postID uuid.UUID, input usecase.UpdatePostInput) (*entity.DiscussionPost, error) {
	post, err := s.postUsecase.UpdatePost(ctx, accountID, postID, input)
	if err != nil {
		return nil, err
	}

	if len(input.AttachmentIds) > 0 {
		if err := s.attachmentUsecase.LinkToPost(ctx, postID, input.AttachmentIds, accountID); err != nil {
			return nil, err
		}
	}

	if input.RemoveAllAttachments {
		attachments, err := s.attachmentUsecase.FindByPostID(ctx, postID)
		if err != nil {
			return nil, err
		}
		if len(attachments) > 0 {
			ids := make([]uuid.UUID, 0, len(attachments))
			for _, att := range attachments {
				ids = append(ids, att.ID)
			}
			if err := s.attachmentUsecase.UnlinkFromPost(ctx, postID, ids, accountID); err != nil {
				return nil, err
			}
		}
	} else if len(input.RemoveAttachmentIds) > 0 {
		if err := s.attachmentUsecase.UnlinkFromPost(ctx, postID, input.RemoveAttachmentIds, accountID); err != nil {
			return nil, err
		}
	}

	return post, nil
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

func (s *communityService) ListThreadFollowStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.followUsecase.ListThreadFollowStatus(ctx, accountID, threadIDs)
}

func (s *communityService) ListThreadUnreadCounts(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	return s.followUsecase.ListThreadUnreadCounts(ctx, accountID, threadIDs)
}

func (s *communityService) ListThreadSolutionStatus(ctx context.Context, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.followUsecase.ListThreadSolutionStatus(ctx, threadIDs)
}

func (s *communityService) MarkThreadRead(ctx context.Context, accountID, threadID uuid.UUID) error {
	return s.followUsecase.MarkThreadRead(ctx, accountID, threadID)
}

func (s *communityService) RecordThreadView(ctx context.Context, accountID, threadID uuid.UUID) error {
	return s.threadRepo.IncrementViews(ctx, threadID)
}

func (s *communityService) NotifyThreadUpdated(ctx context.Context, threadID uuid.UUID) {
	s.wsHub.PublishToThread(threadID, ws.NewThreadUpdatedEvent(threadID.String()))
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
