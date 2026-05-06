package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type DiscussionThreadRepository interface {
	sharedrepo.GenericRepository[entity.DiscussionThread]

	GetBySlug(ctx context.Context, slug string, parentThreadID *uuid.UUID) (*entity.DiscussionThread, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error)
	Search(ctx context.Context, keyword string, categoryID *uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error)
	IncrementViews(ctx context.Context, threadID uuid.UUID) error
	UpdateLastActivity(ctx context.Context, threadID uuid.UUID, at time.Time) error
	UpdateReplyCount(ctx context.Context, threadID uuid.UUID, delta int) error
	GetStatus(ctx context.Context, threadID uuid.UUID) (entity.ThreadStatus, error)
	IsAuthor(ctx context.Context, threadID, accountID uuid.UUID) (bool, error)
}
