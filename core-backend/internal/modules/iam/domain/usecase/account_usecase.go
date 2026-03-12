package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type AccountUsecase interface {
	CreateAccount(ctx context.Context, input CreateAccountInput) (*entity.Account, error)
	GetAccount(ctx context.Context, accountID uuid.UUID) (*entity.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error)
	ListUserAccounts(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error)
	UpdateAccount(ctx context.Context, accountID uuid.UUID, input UpdateAccountInput) (*entity.Account, error)
	ChangeAccountStatus(ctx context.Context, accountID uuid.UUID, status entity.AccountStatus) error
}

type CreateAccountInput struct {
	UserID       uuid.UUID
	Email        string
	PasswordHash string
	PhoneNumber  *string
}

type UpdateAccountInput struct {
	Email       *string
	PhoneNumber *string
}
