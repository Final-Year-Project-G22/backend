package repository

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type userRepository struct {
	sharedrepo.GenericRepository[entity.User]
}

// NewUserRepository creates a new UserRepository implementation.
func NewUserRepository(db *core.Database, logger core.Logger) repository.UserRepository {
	base := sharedrepo.NewBaseRepository[entity.User](db, logger)
	return &userRepository{
		GenericRepository: base,
	}
}
