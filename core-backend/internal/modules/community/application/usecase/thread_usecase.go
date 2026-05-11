package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	iamrepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/taxonomy"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type discussionThreadUsecase struct {
	threadRepo  repository.DiscussionThreadRepository
	postRepo    repository.DiscussionPostRepository
	profileRepo iamrepository.BusinessProfileRepository
	validator   *taxonomy.TaxonomyValidator
	transactor  sharedrepo.Transactor
	logger      core.Logger
}

func NewDiscussionThreadUsecase(
	threadRepo repository.DiscussionThreadRepository,
	postRepo repository.DiscussionPostRepository,
	profileRepo iamrepository.BusinessProfileRepository,
	validator *taxonomy.TaxonomyValidator,
	transactor sharedrepo.Transactor,
	logger core.Logger,
) usecase.DiscussionThreadUsecase {
	return &discussionThreadUsecase{
		threadRepo:  threadRepo,
		postRepo:    postRepo,
		profileRepo: profileRepo,
		validator:   validator,
		transactor:  transactor,
		logger:      logger,
	}
}

func (u *discussionThreadUsecase) CreateThread(ctx context.Context, accountID uuid.UUID, input usecase.CreateThreadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, nil, apperrors.RequiredFieldError("title")
	}
	if strings.TrimSpace(input.Slug) == "" {
		return nil, nil, apperrors.RequiredFieldError("slug")
	}
	if strings.TrimSpace(input.InitialPostContent) == "" {
		return nil, nil, apperrors.RequiredFieldError("initialPostContent")
	}

	if err := u.validator.Validate(ctx, input.SectorIDs, input.TagIDs); err != nil {
		return nil, nil, err
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

	existing, err := u.threadRepo.GetBySlug(ctx, input.Slug, input.ParentThreadID)
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
			SectorIDs:       input.SectorIDs,
			TagIDs:          input.TagIDs,
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

	if thread.Status != entity.ThreadStatusActive {
		return nil, apperrors.BadRequestError("community.errors.threadNotActive")
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
	if input.SectorIDs != nil {
		if err := u.validator.Validate(ctx, *input.SectorIDs, nil); err != nil {
			return nil, err
		}
		if len(*input.SectorIDs) == 0 {
			thread.SectorIDs = nil
		} else {
			thread.SectorIDs = *input.SectorIDs
		}
	}
	if input.TagIDs != nil {
		if err := u.validator.Validate(ctx, nil, *input.TagIDs); err != nil {
			return nil, err
		}
		if len(*input.TagIDs) == 0 {
			thread.TagIDs = nil
		} else {
			thread.TagIDs = *input.TagIDs
		}
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

func (u *discussionThreadUsecase) buildTaxonomyFilter(ctx context.Context, accountID uuid.UUID) (sectorIDs []uuid.UUID, tagIDs []uuid.UUID) {
	profile, err := u.profileRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		u.logger.Error("Failed to get business profile for thread filtering", core.Error(err))
		return nil, nil
	}
	if profile == nil {
		return nil, nil
	}
	if profile.SectorID != nil {
		sectorIDs = append(sectorIDs, *profile.SectorID)
	}
	for _, tag := range profile.Tags {
		tagIDs = append(tagIDs, tag.ID)
	}
	return sectorIDs, tagIDs
}

func (u *discussionThreadUsecase) ListThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	sectorIDs, tagIDs := u.buildTaxonomyFilter(ctx, accountID)
	return u.threadRepo.ListByTaxonomy(ctx, sectorIDs, tagIDs, q)
}

func (u *discussionThreadUsecase) SearchThreads(ctx context.Context, keyword string, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	return u.threadRepo.Search(ctx, keyword, nil, nil, q)
}

func (u *discussionThreadUsecase) ListAllThreads(ctx context.Context, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	return u.threadRepo.ListByTaxonomy(ctx, nil, nil, q)
}

func (u *discussionThreadUsecase) GetThread(ctx context.Context, threadID uuid.UUID) (*entity.DiscussionThread, error) {
	return u.threadRepo.GetByID(ctx, threadID)
}

func (u *discussionThreadUsecase) DeleteThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	thread, err := u.threadRepo.GetByID(ctx, threadID)
	if err != nil {
		return err
	}

	if ok, err := u.threadRepo.IsAuthor(ctx, threadID, accountID); err != nil {
		return err
	} else if !ok {
		return apperrors.ForbiddenError("community.errors.permissionDenied")
	}

	if thread.ReplyCount > 0 {
		return apperrors.BadRequestError("community.errors.threadHasReplies")
	}

	return u.threadRepo.Delete(ctx, threadID)
}
