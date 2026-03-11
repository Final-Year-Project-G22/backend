package appusecase

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type userUsecase struct {
	userRepo repository.UserRepository
	logger   core.Logger
}

func NewUserUsecase(
	userRepo repository.UserRepository,
	logger core.Logger,
) usecase.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *userUsecase) CreateUser(ctx context.Context, input usecase.CreateUserInput) (*entity.User, error) {
	user := &entity.User{
		FirstName: strings.TrimSpace(input.FirstName),
		LastName:  strings.TrimSpace(input.LastName),
		ImageURL:  input.ImageURL,
		Bio:       input.Bio,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	u.logger.Info("User created", core.String("userID", user.ID.String()))
	return user, nil
}

func (u *userUsecase) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}

func (u *userUsecase) ListUsers(ctx context.Context) ([]*entity.User, error) {
	// TODO: Add pagination support
	return u.userRepo.Find(ctx, query.QueryOptions{})
}

func (u *userUsecase) UpdateUser(ctx context.Context, userID uuid.UUID, input usecase.UpdateUserInput) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.FirstName != nil {
		user.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		user.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.ImageURL != nil {
		user.ImageURL = input.ImageURL
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	u.logger.Info("User updated", core.String("userID", user.ID.String()))
	return user, nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := u.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	u.logger.Info("User deleted", core.String("userID", userID.String()))
	return nil
}
