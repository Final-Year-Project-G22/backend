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

type discussionThreadUsecase struct {
	catRepo    repository.CommunityCategoryRepository
	threadRepo repository.DiscussionThreadRepository
	postRepo   repository.DiscussionPostRepository
	transactor sharedrepo.Transactor
}

func NewDiscussionThreadUsecase(
	catRepo repository.CommunityCategoryRepository,
	threadRepo repository.DiscussionThreadRepository,
	postRepo repository.DiscussionPostRepository,
	transactor sharedrepo.Transactor,
) usecase.DiscussionThreadUsecase {
	return &discussionThreadUsecase{
		catRepo:    catRepo,
		threadRepo: threadRepo,
		postRepo:   postRepo,
		transactor: transactor,
	}
}

func (u *discussionThreadUsecase) CreateThread(ctx context.Context, accountID uuid.UUID, input usecase.CreateThreadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error) {
	if input.CategoryID == uuid.Nil {
		return nil, nil, apperrors.RequiredFieldError("categoryId")
	}
	if strings.TrimSpace(input.Title) == "" {
		return nil, nil, apperrors.RequiredFieldError("title")
	}
	if strings.TrimSpace(input.Slug) == "" {
		return nil, nil, apperrors.RequiredFieldError("slug")
	}
	if strings.TrimSpace(input.InitialPostContent) == "" {
		return nil, nil, apperrors.RequiredFieldError("initialPostContent")
	}

	category, err := u.catRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, nil, err
	}
	if !category.IsActive {
		return nil, nil, apperrors.InvalidInputError("categoryId", "community.errors.categoryInactive")
	}

	if input.ParentThreadID != nil {
		parentThread, err := u.threadRepo.GetByID(ctx, *input.ParentThreadID)
		if err != nil {
			if err == communityerror.ErrThreadNotFound {
				return nil, nil, apperrors.InvalidInputError("parentThreadId", "community.errors.threadNotFound")
			}
			return nil, nil, err
		}
		if parentThread.Status != entity.ThreadStatusActive {
			return nil, nil, apperrors.InvalidInputError("parentThreadId", "community.errors.threadClosed")
		}
	}

	existing, err := u.threadRepo.GetBySlug(ctx, input.CategoryID, input.Slug, input.ParentThreadID)
	if err != nil && err != communityerror.ErrThreadNotFound {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, apperrors.AlreadyExistsError("thread", "slug", input.Slug)
	}

	var thread *entity.DiscussionThread
	var post *entity.DiscussionPost
	now := time.Now().UTC()
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		thread = &entity.DiscussionThread{
			CategoryID:      input.CategoryID,
			ParentThreadID:  input.ParentThreadID,
			AuthorAccountID: accountID,
			Title:           strings.TrimSpace(input.Title),
			Slug:            strings.TrimSpace(input.Slug),
			Description:     input.Description,
			Status:          entity.ThreadStatusActive,
			LastActivityAt:  &now,
		}
		if err := u.threadRepo.Create(txCtx, thread); err != nil {
			return err
		}
		post = &entity.DiscussionPost{
			ThreadID:        thread.ID,
			AuthorAccountID: accountID,
			Content:         strings.TrimSpace(input.InitialPostContent),
			AttachmentURL:   input.AttachmentURL,
			AttachmentType:  input.AttachmentType,
		}
		return u.postRepo.Create(txCtx, post)
	}); err != nil {
		return nil, nil, err
	}
	return thread, post, nil
}

func (u *discussionThreadUsecase) UpdateThread(ctx context.Context, accountID, threadID uuid.UUID, input usecase.UpdateThreadInput) (*entity.DiscussionThread, error) {
	thread, err := u.threadRepo.GetByID(ctx, threadID)
	if err != nil {
		return nil, err
	}
	if ok, err := u.threadRepo.IsAuthor(ctx, threadID, accountID); err != nil {
		return nil, err
	} else if !ok {
		return nil, apperrors.ForbiddenError("community.errors.permissionDenied")
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, apperrors.RequiredFieldError("title")
		}
		thread.Title = title
	}
	if input.Description != nil {
		thread.Description = input.Description
	}
	if input.IsPinned != nil {
		thread.IsPinned = *input.IsPinned
	}
	if input.Status != nil {
		thread.Status = *input.Status
	}
	if err := u.threadRepo.Update(ctx, thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (u *discussionThreadUsecase) CloseThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	if ok, err := u.threadRepo.IsAuthor(ctx, threadID, accountID); err != nil {
		return err
	} else if !ok {
		return apperrors.ForbiddenError("community.errors.permissionDenied")
	}
	return u.threadRepo.UpdateByID(ctx, threadID, map[string]interface{}{
		"status":     entity.ThreadStatusClosed,
		"updated_at": time.Now().UTC(),
	})
}

func (u *discussionThreadUsecase) ListThreadsByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	return u.threadRepo.ListByCategory(ctx, categoryID, q)
}

func (u *discussionThreadUsecase) SearchThreads(ctx context.Context, keyword string, categoryID *uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	return u.threadRepo.Search(ctx, keyword, categoryID, q)
}

func (u *discussionThreadUsecase) GetThread(ctx context.Context, threadID uuid.UUID) (*entity.DiscussionThread, error) {
	return u.threadRepo.GetByID(ctx, threadID)
}
