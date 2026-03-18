package repository

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type UserRepository interface {
	sharedrepo.GenericRepository[entity.User]

	UpdateAvatar(userID string, imageURL string) error
	GetImageURL(userID string) (string, error)
}
