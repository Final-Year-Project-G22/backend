package usecase

import (
	"context"

	"github.com/google/uuid"
)

type CommunityFollowUsecase interface {
	FollowThread(ctx context.Context, accountID, threadID uuid.UUID) error
	UnfollowThread(ctx context.Context, accountID, threadID uuid.UUID) error
	FollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
	UnfollowCategory(ctx context.Context, accountID, categoryID uuid.UUID) error
}
