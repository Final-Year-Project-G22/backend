package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	sharedrepo.GenericRepository[entity.User]
	db     *core.Database
	logger core.Logger
}

func NewUserRepository(db *core.Database, logger core.Logger) repository.UserRepository {
	base := sharedrepo.NewBaseRepository[entity.User](db, logger)
	return &userRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *userRepository) UpdateAvatar(userID string, imageURL string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequestError("invalid user id")
	}

	if err := r.db.WithContext(context.Background()).Model(new(entity.User)).Where("id = ?", id).Update("image_url", imageURL).Error; err != nil {
		r.logger.Error("Failed to update user avatar", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userRepository) GetImageURL(userID string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", errors.BadRequestError("invalid user id")
	}

	var user entity.User
	if err := r.db.WithContext(context.Background()).Select("image_url").First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		r.logger.Error("Failed to get user image URL", core.Error(err))
		return "", errors.InternalError("errors.databaseError", err)
	}

	if user.ImageURL == nil {
		return "", nil
	}
	return *user.ImageURL, nil
}
