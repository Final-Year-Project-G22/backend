package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type DiscussionPostRepository interface {
	sharedrepo.GenericRepository[entity.DiscussionPost]

	ListByThread(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionPost, error)
	ListReplies(ctx context.Context, parentPostID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionPost, error)
	GetSolution(ctx context.Context, threadID uuid.UUID) (*entity.DiscussionPost, error)
	ClearSolution(ctx context.Context, threadID uuid.UUID) error
	IsAuthor(ctx context.Context, postID, accountID uuid.UUID) (bool, error)
	CountUnreadByThreadIDs(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]int, error)
	ListSolutionStatus(ctx context.Context, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}
