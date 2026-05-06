package repository

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

// TagRepository provides access to tag storage.
type TagRepository interface {
	sharedrepo.GenericRepository[entity.Tag]
}

// SectorRepository provides access to sector storage.
type SectorRepository interface {
	sharedrepo.GenericRepository[entity.Sector]
}
