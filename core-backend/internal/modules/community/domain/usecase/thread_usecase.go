package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type DiscussionThreadUsecase interface {
	CreateThread(ctx context.Context, accountID uuid.UUID, input CreateThreadInput) (*entity.DiscussionThread, *entity.DiscussionPost, error)
	UpdateThread(ctx context.Context, accountID, threadID uuid.UUID, input UpdateThreadInput) (*entity.DiscussionThread, error)
	CloseThread(ctx context.Context, accountID, threadID uuid.UUID) error
	ListThreads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error)
	SearchThreads(ctx context.Context, keyword string, q query.QueryOptions) ([]*entity.DiscussionThread, error)
	ListAllThreads(ctx context.Context, q query.QueryOptions) ([]*entity.DiscussionThread, error)
	GetThread(ctx context.Context, threadID uuid.UUID) (*entity.DiscussionThread, error)
	DeleteThread(ctx context.Context, accountID, threadID uuid.UUID) error
}
