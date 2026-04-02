package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ProgressRepository interface {
	sharedrepo.GenericRepository[entity.UserGuideProgress]

	GetProgress(ctx context.Context, accountID, userID, stepID uuid.UUID) (*entity.UserGuideProgress, error)
	ListProgressByGuide(ctx context.Context, accountID, userID, guideID uuid.UUID, q query.QueryOptions) ([]*entity.UserGuideProgress, error)
	UpsertProgress(ctx context.Context, progress *entity.UserGuideProgress) error
	DeleteProgress(ctx context.Context, accountID, userID, stepID uuid.UUID) error

	GetJourney(ctx context.Context, accountID, userID, guideID uuid.UUID) (*entity.UserGuideJourney, error)
	UpsertJourney(ctx context.Context, journey *entity.UserGuideJourney) error
	DeleteJourney(ctx context.Context, accountID, userID, guideID uuid.UUID) error
	InvalidateJourneysForGuide(ctx context.Context, guideID uuid.UUID) error
	InvalidateJourneyForUser(ctx context.Context, accountID, userID, guideID uuid.UUID) error

	GetBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID) (*entity.UserGuideBookmark, error)
	ListBookmarks(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions) ([]*entity.UserGuideBookmark, error)
	UpsertBookmark(ctx context.Context, bookmark *entity.UserGuideBookmark) error
	RemoveBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID) error

	UpsertRecentView(ctx context.Context, accountID, userID, guideID uuid.UUID) error
	ListRecentlyViewedGuides(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*entity.Guide, error)
}
