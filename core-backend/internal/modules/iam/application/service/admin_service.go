package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/validation"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/event"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const adminPasswordLength = 12

type AdminService interface {
	RegisterAdmin(ctx context.Context, input RegisterAdminInput) (*RegisterAdminOutput, error)
	UpdateAdminRoles(ctx context.Context, input UpdateAdminRolesInput) error
}

type RegisterAdminInput struct {
	Email      string
	Username   *string
	FirstName  string
	LastName   string
	RoleIDs    []uuid.UUID
	AssignedBy uuid.UUID
}

type RegisterAdminOutput struct {
	AccountID uuid.UUID
}

type UpdateAdminRolesInput struct {
	AccountID uuid.UUID
	RoleIDs   []uuid.UUID
	UpdatedBy uuid.UUID
}

type adminService struct {
	transactor         sharedrepo.Transactor
	userUsecase        usecase.UserUsecase
	accountUsecase     usecase.AccountUsecase
	roleRepo           repository.RoleRepository
	roleAssignmentRepo repository.RoleAssignmentRepository
	messageBus         rabbitmq.Bus
	logger             core.Logger
}

func NewAdminService(
	transactor sharedrepo.Transactor,
	userUsecase usecase.UserUsecase,
	accountUsecase usecase.AccountUsecase,
	roleRepo repository.RoleRepository,
	roleAssignmentRepo repository.RoleAssignmentRepository,
	messageBus rabbitmq.Bus,
	logger core.Logger,
) AdminService {
	return &adminService{
		transactor:         transactor,
		userUsecase:        userUsecase,
		accountUsecase:     accountUsecase,
		roleRepo:           roleRepo,
		roleAssignmentRepo: roleAssignmentRepo,
		messageBus:         messageBus,
		logger:             logger,
	}
}

func (s *adminService) RegisterAdmin(ctx context.Context, input RegisterAdminInput) (*RegisterAdminOutput, error) {
	email, ok := validation.NormalizeEmail(input.Email)
	if !ok {
		return nil, errors.BadRequestError("iam.errors.invalidEmailFormat")
	}

	if input.Username != nil && strings.TrimSpace(*input.Username) != "" {
		if _, valid := validation.NormalizeUsername(*input.Username); !valid {
			return nil, errors.BadRequestError("iam.errors.invalidUsernameFormat")
		}
	}

	if existingAccount, err := s.accountUsecase.GetAccountByEmail(ctx, email); err == nil && existingAccount != nil {
		return nil, errors.ConflictError("iam.errors.emailAlreadyExists")
	} else if err != nil && err != iamerror.ErrAccountNotFound {
		return nil, err
	}

	if len(input.RoleIDs) == 0 {
		return nil, errors.BadRequestError("iam.errors.invalidInput")
	}

	roles, err := s.roleRepo.FindByIDs(ctx, input.RoleIDs)
	if err != nil {
		return nil, err
	}

	if len(roles) != len(uniqueIDs(input.RoleIDs)) {
		return nil, errors.NotFoundErrorWithKey("iam.errors.notFound")
	}

	password, err := generateAdminPassword(adminPasswordLength)
	if err != nil {
		return nil, errors.InternalError("iam.errors.failedToGenerateOtp", err)
	}

	var account *entity.Account
	var user *entity.User

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		var txErr error
		user, txErr = s.userUsecase.CreateUser(txCtx, usecase.CreateUserInput{
			FirstName: input.FirstName,
			LastName:  input.LastName,
		})
		if txErr != nil {
			return txErr
		}

		passwordHash, txErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if txErr != nil {
			return errors.InternalError("iam.errors.passwordHashFailed", txErr)
		}
		passwordHashStr := string(passwordHash)

		account, txErr = s.accountUsecase.CreateAccount(txCtx, usecase.CreateAccountInput{
			UserID:        user.ID,
			Email:         email,
			Username:      input.Username,
			PasswordHash:  &passwordHashStr,
			EmailVerified: true,
			Status:        entity.AccountStatusActive,
		})
		if txErr != nil {
			return txErr
		}

		for _, role := range roles {
			assignment := &entity.RoleAssignment{
				AccountID:  account.ID,
				RoleID:     role.ID,
				AssignedBy: input.AssignedBy,
			}
			if txErr := s.roleAssignmentRepo.Create(txCtx, assignment); txErr != nil {
				return txErr
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go s.publishAdminCreated(context.Background(), account, user, password)

	return &RegisterAdminOutput{AccountID: account.ID}, nil
}

func (s *adminService) UpdateAdminRoles(ctx context.Context, input UpdateAdminRolesInput) error {
	if len(input.RoleIDs) == 0 {
		return errors.BadRequestError("iam.errors.invalidInput")
	}

	if _, err := s.accountUsecase.GetAccount(ctx, input.AccountID); err != nil {
		return err
	}

	roles, err := s.roleRepo.FindByIDs(ctx, input.RoleIDs)
	if err != nil {
		return err
	}
	if len(roles) != len(uniqueIDs(input.RoleIDs)) {
		return errors.NotFoundErrorWithKey("iam.errors.notFound")
	}

	return s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		assignments, err := s.roleAssignmentRepo.ListByAccountID(txCtx, input.AccountID)
		if err != nil {
			return err
		}
		for _, assignment := range assignments {
			if err := s.roleAssignmentRepo.Revoke(txCtx, assignment.ID, time.Now(), stringPtr("roles updated")); err != nil {
				return err
			}
		}

		for _, role := range roles {
			assignment := &entity.RoleAssignment{
				AccountID:  input.AccountID,
				RoleID:     role.ID,
				AssignedBy: input.UpdatedBy,
			}
			if err := s.roleAssignmentRepo.Create(txCtx, assignment); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *adminService) publishAdminCreated(ctx context.Context, account *entity.Account, user *entity.User, password string) {
	err := s.messageBus.Publish(ctx, event.AdminCreated, event.AdminCreatedEvent{
		AccountID: account.ID.String(),
		Email:     account.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  password,
		Locale:    i18n.LocaleFromContext(ctx),
	})
	if err != nil {
		s.logger.Error("Failed to publish admin credentials event", core.Error(err))
	}
}

func generateAdminPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	maxIndex := big.NewInt(int64(len(charset)))

	var builder strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", err
		}
		builder.WriteByte(charset[n.Int64()])
	}

	return builder.String(), nil
}

func uniqueIDs(ids []uuid.UUID) map[uuid.UUID]struct{} {
	unique := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	return unique
}
