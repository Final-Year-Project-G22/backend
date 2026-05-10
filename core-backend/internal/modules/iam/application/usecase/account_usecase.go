package appusecase

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/validation"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
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
	normalizedEmail, ok := validation.NormalizeEmail(input.Email)
	if !ok {
		return nil, errors.BadRequestError("iam.errors.invalidEmailFormat")
	}

	var username *string
	var usernameNormalized *string
	if input.Username != nil {
		if strings.TrimSpace(*input.Username) != "" {
			normalized, valid := validation.NormalizeUsername(*input.Username)
			if !valid {
				return nil, errors.BadRequestError("iam.errors.invalidUsernameFormat")
			}
			exists, err := u.accountRepo.ExistsByUsernameNormalized(ctx, normalized)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.ConflictError("iam.errors.usernameAlreadyExists")
			}

			username = &normalized
			usernameNormalized = &normalized
		}
	}

	status := input.Status
	if status == "" {
		status = entity.AccountStatusPendingVerification
	}

	account := &entity.Account{
		UserID:             input.UserID,
		Email:              normalizedEmail,
		EmailNormalized:    normalizedEmail,
		Username:           username,
		UsernameNormalized: usernameNormalized,
		PasswordHash:       input.PasswordHash,
		PhoneNumber:        input.PhoneNumber,
		EmailVerified:      input.EmailVerified,
		Status:             status,
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

func (u *accountUsecase) GetAccountByIdentifier(ctx context.Context, identifier string) (*entity.Account, error) {
	return u.accountRepo.GetByEmailOrUsername(ctx, identifier)
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
		normalizedEmail, ok := validation.NormalizeEmail(*input.Email)
		if !ok {
			return nil, errors.BadRequestError("iam.errors.invalidEmailFormat")
		}
		account.Email = normalizedEmail
		account.EmailNormalized = normalizedEmail
	}
	if input.PhoneNumber != nil {
		account.PhoneNumber = input.PhoneNumber
	}
	if input.LastLoginAt != nil {
		account.LastLoginAt = input.LastLoginAt
	}

	if err := u.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	u.logger.Info("Account updated", core.String("accountID", account.ID.String()))
	return account, nil
}
func (u *accountUsecase) UpdateAccountPassword(ctx context.Context, accountId uuid.UUID, input usecase.UpdateAccountPasswordInput) error {
	account, err := u.accountRepo.GetByID(ctx, accountId)
	if err != nil {
		return err
	}

	hashedPassword := strings.TrimSpace(input.NewHashedPassword)
	if hashedPassword != "" {
		account.PasswordHash = &hashedPassword
	}

	err = u.accountRepo.Update(ctx, account)
	if err != nil {
		return err
	}
	u.logger.Info("Account password updated", core.String("accountId", account.ID.String()))
	return nil

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

func (u *accountUsecase) MarkEmailVerifiedAndActivate(ctx context.Context, accountID uuid.UUID) error {
	if err := u.accountRepo.MarkEmailVerifiedAndActivate(ctx, accountID); err != nil {
		return err
	}

	u.logger.Info("Account email verified and activated",
		core.String("accountID", accountID.String()),
	)

	return nil
}

func (u *accountUsecase) ListAdmins(ctx context.Context, permissionCodes []string, queryOpts map[string]interface{}) ([]*entity.Account, int64, error) {
	return u.accountRepo.ListAdmins(ctx, permissionCodes, queryOpts)
}
