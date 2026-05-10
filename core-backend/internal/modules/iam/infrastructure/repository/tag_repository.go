package repository

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type tagRepository struct {
	sharedrepo.GenericRepository[entity.Tag]
	db     *core.Database
	logger core.Logger
}

// NewTagRepository creates a new TagRepository implementation.
func NewTagRepository(db *core.Database, logger core.Logger) repository.TagRepository {
	base := sharedrepo.NewBaseRepository[entity.Tag](db, logger)
	return &tagRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

type sectorRepository struct {
	sharedrepo.GenericRepository[entity.Sector]
	db     *core.Database
	logger core.Logger
}

// NewSectorRepository creates a new SectorRepository implementation.
func NewSectorRepository(db *core.Database, logger core.Logger) repository.SectorRepository {
	base := sharedrepo.NewBaseRepository[entity.Sector](db, logger)
	return &sectorRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}
