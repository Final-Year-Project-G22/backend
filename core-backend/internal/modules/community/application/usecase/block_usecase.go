package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type threadBlockUsecase struct {
	threadRepo         repository.DiscussionThreadRepository
	blockRepo          repository.ThreadBlockedUserRepository
	threadSettingsRepo repository.UserThreadSettingsRepository
}

func NewThreadBlockUsecase(
	threadRepo repository.DiscussionThreadRepository,
	blockRepo repository.ThreadBlockedUserRepository,
	threadSettingsRepo repository.UserThreadSettingsRepository,
) usecase.ThreadBlockUsecase {
	return &threadBlockUsecase{
		threadRepo:         threadRepo,
		blockRepo:          blockRepo,
		threadSettingsRepo: threadSettingsRepo,
	}
}

func (u *threadBlockUsecase) BlockUser(ctx context.Context, input usecase.BlockUserInput) error {
	if input.ActorID == input.BlockedID {
		return apperrors.InvalidInputError("blockedId", "community.errors.invalidBlockedUser")
	}
	if _, err := u.threadRepo.GetByID(ctx, input.ThreadID); err != nil {
		return err
	}
	if !input.IsAdmin {
		isAuthor, err := u.threadRepo.IsAuthor(ctx, input.ThreadID, input.ActorID)
		if err != nil {
			return err
		}
		if !isAuthor {
			return apperrors.ForbiddenError("community.errors.permissionDenied")
		}
	}
	if err := u.blockRepo.Block(ctx, input.ThreadID, input.BlockedID, input.ActorID, input.Reason); err != nil {
		return err
	}
	if err := u.threadSettingsRepo.Delete(ctx, input.BlockedID, input.ThreadID); err != nil {
		if err == communityerror.ErrThreadSettingNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (u *threadBlockUsecase) UnblockUser(ctx context.Context, input usecase.BlockUserInput) error {
	if _, err := u.threadRepo.GetByID(ctx, input.ThreadID); err != nil {
		return err
	}
	if !input.IsAdmin {
		isAuthor, err := u.threadRepo.IsAuthor(ctx, input.ThreadID, input.ActorID)
		if err != nil {
			return err
		}
		if !isAuthor {
			return apperrors.ForbiddenError("community.errors.permissionDenied")
		}
	}
	return u.blockRepo.Unblock(ctx, input.ThreadID, input.BlockedID)
}

func (u *threadBlockUsecase) IsBlocked(ctx context.Context, threadID, accountID uuid.UUID) (bool, error) {
	return u.blockRepo.IsBlocked(ctx, threadID, accountID)
}

func (u *threadBlockUsecase) ListBlockedUsers(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.ThreadBlockedUser, int64, error) {
	return u.blockRepo.ListBlocked(ctx, threadID, q)
}

func (u *threadBlockUsecase) ListAllBlockedUsers(ctx context.Context, q query.QueryOptions) ([]*entity.ThreadBlockedUser, int64, error) {
	return u.blockRepo.ListAllBlocked(ctx, q)
}
