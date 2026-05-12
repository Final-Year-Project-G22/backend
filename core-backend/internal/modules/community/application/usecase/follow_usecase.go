package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type communityFollowUsecase struct {
	catRepo              repository.CommunityCategoryRepository
	threadRepo           repository.DiscussionThreadRepository
	postRepo             repository.DiscussionPostRepository
	threadSettingsRepo   repository.UserThreadSettingsRepository
	categorySettingsRepo repository.UserCategorySettingsRepository
	blockRepo            repository.ThreadBlockedUserRepository
}

func NewCommunityFollowUsecase(
	catRepo repository.CommunityCategoryRepository,
	threadRepo repository.DiscussionThreadRepository,
	postRepo repository.DiscussionPostRepository,
	threadSettingsRepo repository.UserThreadSettingsRepository,
	categorySettingsRepo repository.UserCategorySettingsRepository,
	blockRepo repository.ThreadBlockedUserRepository,
) usecase.CommunityFollowUsecase {
	return &communityFollowUsecase{
		catRepo:              catRepo,
		threadRepo:           threadRepo,
		postRepo:             postRepo,
		threadSettingsRepo:   threadSettingsRepo,
		categorySettingsRepo: categorySettingsRepo,
		blockRepo:            blockRepo,
	}
}

func (u *communityFollowUsecase) FollowThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	if _, err := u.threadRepo.GetByID(ctx, threadID); err != nil {
		return err
	}
	blocked, err := u.blockRepo.IsBlocked(ctx, threadID, accountID)
	if err != nil {
		return err
	}
	if blocked {
		return communityerror.ErrThreadBlocked
	}
	return u.threadSettingsRepo.UpsertFollow(ctx, accountID, threadID, true)
}

func (u *communityFollowUsecase) UnfollowThread(ctx context.Context, accountID, threadID uuid.UUID) error {
	if err := u.threadSettingsRepo.Delete(ctx, accountID, threadID); err != nil {
		if err == communityerror.ErrThreadSettingNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (u *communityFollowUsecase) FollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error {
	if err := u.ensureCategoryActive(ctx, categoryID); err != nil {
		return err
	}
	return u.categorySettingsRepo.UpsertFollow(ctx, accountID, categoryID, true)
}

func (u *communityFollowUsecase) UnfollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error {
	if err := u.categorySettingsRepo.Delete(ctx, accountID, categoryID); err != nil {
		if err == communityerror.ErrCategorySettingNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (u *communityFollowUsecase) ListFollowedThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error) {
	return u.threadSettingsRepo.ListFollowed(ctx, accountID, q)
}

func (u *communityFollowUsecase) ListFollowedCategories(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error) {
	return u.categorySettingsRepo.ListFollowed(ctx, accountID, q)
}

func (u *communityFollowUsecase) ListThreadFollowStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return u.threadSettingsRepo.ListFollowStatus(ctx, accountID, threadIDs)
}

func (u *communityFollowUsecase) ListThreadUnreadCounts(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	return u.postRepo.CountUnreadByThreadIDs(ctx, accountID, threadIDs)
}

func (u *communityFollowUsecase) MarkThreadRead(ctx context.Context, accountID, threadID uuid.UUID) error {
	return u.threadSettingsRepo.UpdateLastRead(ctx, accountID, threadID, time.Now().UTC())
}

func (u *communityFollowUsecase) ensureCategoryActive(ctx context.Context, categoryID uuid.UUID) error {
	category, err := u.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if !category.IsActive {
		return apperrors.InvalidInputError("categoryId", "community.errors.categoryInactive")
	}
	return nil
}
