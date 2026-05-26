package service

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
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
	MuteThread(ctx context.Context, accountID, threadID uuid.UUID) error
	UnmuteThread(ctx context.Context, accountID, threadID uuid.UUID) error
	FollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	UnfollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	ListFollowedThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error)
	ListFollowedCategories(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error)
	ListThreadFollowStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	ListThreadMuteStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	ListThreadFollowers(ctx context.Context, threadID uuid.UUID) ([]uuid.UUID, error)
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
	db                *core.Database
	threadUsecase     usecase.DiscussionThreadUsecase
	postUsecase       usecase.DiscussionPostUsecase
	attachmentUsecase usecase.AttachmentUsecase
	blockUsecase      usecase.ThreadBlockUsecase
	followUsecase     usecase.CommunityFollowUsecase
	reportUsecase     usecase.ContentReportUsecase
	threadRepo        repository.DiscussionThreadRepository
	wsHub             *ws.Hub
	notifPublisher    NotificationEventPublisher
}

func NewCommunityService(
	db *core.Database,
	threadUsecase usecase.DiscussionThreadUsecase,
	postUsecase usecase.DiscussionPostUsecase,
	attachmentUsecase usecase.AttachmentUsecase,
	blockUsecase usecase.ThreadBlockUsecase,
	followUsecase usecase.CommunityFollowUsecase,
	reportUsecase usecase.ContentReportUsecase,
	threadRepo repository.DiscussionThreadRepository,
	wsHub *ws.Hub,
	notifPublisher NotificationEventPublisher,
) CommunityService {
	return &communityService{
		db:                db,
		threadUsecase:     threadUsecase,
		postUsecase:       postUsecase,
		attachmentUsecase: attachmentUsecase,
		blockUsecase:      blockUsecase,
		followUsecase:     followUsecase,
		reportUsecase:     reportUsecase,
		threadRepo:        threadRepo,
		wsHub:             wsHub,
		notifPublisher:    notifPublisher,
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

	err := s.db.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		if parentPostID != nil {
			post, err = s.postUsecase.ReplyToPost(txCtx, accountID, threadID, *parentPostID, input)
		} else {
			post, err = s.postUsecase.CreatePost(txCtx, accountID, threadID, input)
		}
		if err != nil {
			return err
		}

		if len(input.AttachmentIds) > 0 {
			if err := s.attachmentUsecase.LinkToPost(txCtx, post.ID, input.AttachmentIds, accountID); err != nil {
				return err
			}
		}

		if parentPostID != nil {
			parentPost, err := s.postUsecase.GetPost(txCtx, *parentPostID)
			if err != nil {
				return err
			}

			if parentPost.AuthorAccountID != accountID {
				if err := s.notifPublisher.PublishThreadReply(
					txCtx,
					parentPost.AuthorAccountID,
					threadID,
					post.ID,
					accountID,
				); err != nil {
					return err
				}
			}
		}

		isOwner, err := s.threadRepo.IsAuthor(txCtx, threadID, accountID)
		if err == nil && !isOwner {
			_ = s.followUsecase.FollowThread(txCtx, accountID, threadID)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.wsHub.PublishToThread(threadID, ws.NewPostCreatedEvent(threadID.String(), post.ID.String()))
	s.wsHub.PublishToAll(ws.NewPostCreatedEvent(threadID.String(), post.ID.String()))

	return post, nil
}

func (s *communityService) MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error {
	return s.db.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.postUsecase.MarkSolution(txCtx, accountID, threadID, postID); err != nil {
			return err
		}

		followers, err := s.followUsecase.ListThreadFollowers(txCtx, threadID)
		if err != nil {
			return err
		}

		for _, followerID := range followers {
			if followerID == accountID {
				continue
			}

			_ = s.notifPublisher.PublishThreadSolution(
				txCtx,
				followerID,
				threadID,
				postID,
				accountID,
			)
		}

		return nil
	})
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

func (s *communityService) MuteThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	return s.followUsecase.MuteThread(ctx, accountID, threadID)
}

func (s *communityService) UnmuteThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	return s.followUsecase.UnmuteThread(ctx, accountID, threadID)
}

func (s *communityService) ListThreadMuteStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.followUsecase.ListThreadMuteStatus(ctx, accountID, threadIDs)
}

func (s *communityService) ListThreadFollowers(ctx context.Context, threadID uuid.UUID) ([]uuid.UUID, error) {
	return s.followUsecase.ListThreadFollowers(ctx, threadID)
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
