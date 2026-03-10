package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*entity.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	ListUsers(ctx context.Context) ([]*entity.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*entity.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type CreateUserInput struct {
	FirstName string
	LastName  string
	ImageURL  *string
	Bio       *string
}

type UpdateUserInput struct {
	FirstName *string
	LastName  *string
	ImageURL  *string
	Bio       *string
}
