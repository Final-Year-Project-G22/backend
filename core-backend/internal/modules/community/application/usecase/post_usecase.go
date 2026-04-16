package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

const (
	postEditWindow   = 15 * time.Minute
	postDeleteWindow = 30 * time.Minute
)

type discussionPostUsecase struct {
	threadRepo repository.DiscussionThreadRepository
	postRepo   repository.DiscussionPostRepository
	blockRepo  repository.ThreadBlockedUserRepository
	transactor sharedrepo.Transactor
}

func NewDiscussionPostUsecase(
	threadRepo repository.DiscussionThreadRepository,
	postRepo repository.DiscussionPostRepository,
	blockRepo repository.ThreadBlockedUserRepository,
	transactor sharedrepo.Transactor,
) usecase.DiscussionPostUsecase {
	return &discussionPostUsecase{
		threadRepo: threadRepo,
		postRepo:   postRepo,
		blockRepo:  blockRepo,
		transactor: transactor,
	}
}

func (u *discussionPostUsecase) GetPost(ctx context.Context, postID uuid.UUID) (*entity.DiscussionPost, error) {
	return u.postRepo.GetByID(ctx, postID)
}

func (u *discussionPostUsecase) ListPostsByThread(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionPost, error) {
	if _, err := u.threadRepo.GetByID(ctx, threadID); err != nil {
		return nil, err
	}
	return u.postRepo.ListByThread(ctx, threadID, q)
}

func (u *discussionPostUsecase) CreatePost(ctx context.Context, accountID, threadID uuid.UUID, input usecase.CreatePostInput) (*entity.DiscussionPost, error) {
	if strings.TrimSpace(input.Content) == "" {
		return nil, apperrors.RequiredFieldError("content")
	}
	status, err := u.threadRepo.GetStatus(ctx, threadID)
	if err != nil {
		if err == communityerror.ErrThreadNotFound {
			return nil, apperrors.NotFoundError("thread", threadID)
		}
		return nil, err
	}
	if status != entity.ThreadStatusActive {
		return nil, apperrors.BadRequestError("community.errors.threadInactive")
	}
	blocked, err := u.blockRepo.IsBlocked(ctx, threadID, accountID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, apperrors.ForbiddenError("community.errors.threadBlocked")
	}

	var post *entity.DiscussionPost
	now := time.Now().UTC()
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		post = &entity.DiscussionPost{
			ThreadID:        threadID,
			AuthorAccountID: accountID,
			Content:         strings.TrimSpace(input.Content),
			AttachmentURL:   input.AttachmentURL,
			AttachmentType:  input.AttachmentType,
		}
		if err := u.postRepo.Create(txCtx, post); err != nil {
			return err
		}
		if err := u.threadRepo.UpdateReplyCount(txCtx, threadID, 1); err != nil {
			return err
		}
		return u.threadRepo.UpdateLastActivity(txCtx, threadID, now)
	}); err != nil {
		return nil, err
	}
	return post, nil
}

func (u *discussionPostUsecase) ReplyToPost(ctx context.Context, accountID, threadID, parentPostID uuid.UUID, input usecase.CreatePostInput) (*entity.DiscussionPost, error) {
	if strings.TrimSpace(input.Content) == "" {
		return nil, apperrors.RequiredFieldError("content")
	}
	parent, err := u.postRepo.GetByID(ctx, parentPostID)
	if err != nil {
		return nil, err
	}
	if parent.ThreadID != threadID {
		return nil, apperrors.InvalidInputError("parentPostId", "community.errors.invalidParentPost")
	}
	status, err := u.threadRepo.GetStatus(ctx, threadID)
	if err != nil {
		if err == communityerror.ErrThreadNotFound {
			return nil, apperrors.NotFoundError("thread", threadID)
		}
		return nil, err
	}
	if status != entity.ThreadStatusActive {
		return nil, apperrors.BadRequestError("community.errors.threadInactive")
	}
	blocked, err := u.blockRepo.IsBlocked(ctx, threadID, accountID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, apperrors.ForbiddenError("community.errors.threadBlocked")
	}

	var post *entity.DiscussionPost
	now := time.Now().UTC()
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		post = &entity.DiscussionPost{
			ThreadID:        threadID,
			ParentPostID:    &parentPostID,
			AuthorAccountID: accountID,
			Content:         strings.TrimSpace(input.Content),
			AttachmentURL:   input.AttachmentURL,
			AttachmentType:  input.AttachmentType,
		}
		if err := u.postRepo.Create(txCtx, post); err != nil {
			return err
		}
		if err := u.threadRepo.UpdateReplyCount(txCtx, threadID, 1); err != nil {
			return err
		}
		return u.threadRepo.UpdateLastActivity(txCtx, threadID, now)
	}); err != nil {
		return nil, err
	}
	return post, nil
}

func (u *discussionPostUsecase) UpdatePost(ctx context.Context, accountID, postID uuid.UUID, input usecase.UpdatePostInput) (*entity.DiscussionPost, error) {
	post, err := u.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if ok, err := u.postRepo.IsAuthor(ctx, postID, accountID); err != nil {
		return nil, err
	} else if !ok {
		return nil, apperrors.ForbiddenError("community.errors.permissionDenied")
	}
	if post.IsSolution {
		return nil, apperrors.ForbiddenError("community.errors.postLocked")
	}
	if post.CreatedAt != nil {
		if time.Since(*post.CreatedAt) > postEditWindow {
			return nil, apperrors.ForbiddenError("community.errors.editWindowExpired")
		}
	}
	if solution, err := u.postRepo.GetSolution(ctx, post.ThreadID); err != nil {
		return nil, err
	} else if solution != nil {
		return nil, apperrors.ForbiddenError("community.errors.postLocked")
	}
	updated := false
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, apperrors.RequiredFieldError("content")
		}
		post.Content = content
		updated = true
	}
	if input.AttachmentURL != nil {
		post.AttachmentURL = input.AttachmentURL
		updated = true
	}
	if input.AttachmentType != nil {
		post.AttachmentType = input.AttachmentType
		updated = true
	}
	if input.RemoveAttachment != nil && *input.RemoveAttachment {
		updated = true
		post.AttachmentURL = nil
		post.AttachmentType = nil
	}
	if !updated {
		return nil, apperrors.BadRequestError("community.errors.noChanges")
	}
	post.EditCount++
	now := time.Now().UTC()
	post.EditedAt = &now
	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (u *discussionPostUsecase) DeletePost(ctx context.Context, accountID, postID uuid.UUID) error {
	post, err := u.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if ok, err := u.postRepo.IsAuthor(ctx, postID, accountID); err != nil {
		return err
	} else if !ok {
		return apperrors.ForbiddenError("community.errors.permissionDenied")
	}
	if post.IsSolution {
		return apperrors.ForbiddenError("community.errors.postLocked")
	}
	if post.CreatedAt != nil {
		if time.Since(*post.CreatedAt) > postDeleteWindow {
			return apperrors.ForbiddenError("community.errors.deleteWindowExpired")
		}
	}
	if solution, err := u.postRepo.GetSolution(ctx, post.ThreadID); err != nil {
		return err
	} else if solution != nil {
		return apperrors.ForbiddenError("community.errors.postLocked")
	}
	return u.postRepo.Delete(ctx, postID)
}

func (u *discussionPostUsecase) MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error {
	if ok, err := u.threadRepo.IsAuthor(ctx, threadID, accountID); err != nil {
		return err
	} else if !ok {
		return apperrors.ForbiddenError("community.errors.permissionDenied")
	}
	post, err := u.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.ThreadID != threadID {
		return apperrors.InvalidInputError("postId", "community.errors.invalidPost")
	}
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.postRepo.ClearSolution(txCtx, threadID); err != nil {
			return err
		}
		post.IsSolution = true
		return u.postRepo.Update(txCtx, post)
	})
}
