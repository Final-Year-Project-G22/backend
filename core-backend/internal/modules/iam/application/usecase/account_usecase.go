package appusecase

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/google/uuid"
)

type accountUsecase struct {
	accountRepo repository.AccountRepository
	logger      core.Logger
}

func NewAccountUsecase(
	accountRepo repository.AccountRepository,
	logger core.Logger,
) usecase.AccountUsecase {
	return &accountUsecase{
		accountRepo: accountRepo,
		logger:      logger,
	}
}

func (u *accountUsecase) CreateAccount(ctx context.Context, input usecase.CreateAccountInput) (*entity.Account, error) {
	email := strings.TrimSpace(input.Email)
	normalizedEmail := strings.ToLower(email)

	account := &entity.Account{
		UserID:          input.UserID,
		Email:           email,
		EmailNormalized: normalizedEmail,
		PasswordHash:    input.PasswordHash,
		PhoneNumber:     input.PhoneNumber,
		Status:          entity.AccountStatusPendingVerification,
	}

	if err := u.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	u.logger.Info("Account created",
		core.String("accountID", account.ID.String()),
	)
	return account, nil
}

func (u *accountUsecase) GetAccount(ctx context.Context, accountID uuid.UUID) (*entity.Account, error) {
	return u.accountRepo.GetByID(ctx, accountID)
}

func (u *accountUsecase) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
	return u.accountRepo.GetByEmailNormalized(ctx, email)
}

func (u *accountUsecase) ListUserAccounts(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	return u.accountRepo.ListByUserID(ctx, userID)
}

func (u *accountUsecase) UpdateAccount(ctx context.Context, accountID uuid.UUID, input usecase.UpdateAccountInput) (*entity.Account, error) {
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)
		account.Email = email
		account.EmailNormalized = strings.ToLower(email)
	}
	if input.PhoneNumber != nil {
		account.PhoneNumber = input.PhoneNumber
	}

	if err := u.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	u.logger.Info("Account updated", core.String("accountID", account.ID.String()))
	return account, nil
}

func (u *accountUsecase) ChangeAccountStatus(ctx context.Context, accountID uuid.UUID, status entity.AccountStatus) error {
	if err := u.accountRepo.UpdateStatus(ctx, accountID, status); err != nil {
		return err
	}

	u.logger.Info("Account status changed",
		core.String("accountID", accountID.String()),
		core.String("status", string(status)),
	)
	return nil
}
