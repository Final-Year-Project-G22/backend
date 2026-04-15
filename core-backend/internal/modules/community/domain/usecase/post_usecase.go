package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type DiscussionPostUsecase interface {
	GetPost(ctx context.Context, postID uuid.UUID) (*entity.DiscussionPost, error)
	ListPostsByThread(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionPost, error)
	CreatePost(ctx context.Context, accountID, threadID uuid.UUID, input CreatePostInput) (*entity.DiscussionPost, error)
	ReplyToPost(ctx context.Context, accountID, threadID, parentPostID uuid.UUID, input CreatePostInput) (*entity.DiscussionPost, error)
	UpdatePost(ctx context.Context, accountID, postID uuid.UUID, input UpdatePostInput) (*entity.DiscussionPost, error)
	DeletePost(ctx context.Context, accountID, postID uuid.UUID) error
	MarkSolution(ctx context.Context, accountID, threadID, postID uuid.UUID) error
}
