package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CommunityFollowUsecase interface {
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
}
