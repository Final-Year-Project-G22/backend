package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type TagRepository interface {
	sharedrepo.GenericRepository[entity.Tag]
	UpsertTranslation(ctx context.Context, translation *entity.TagTranslation) error
}

type SectorRepository interface {
	sharedrepo.GenericRepository[entity.Sector]
	UpsertTranslation(ctx context.Context, translation *entity.SectorTranslation) error
}
