package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/validation"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	superAdminEmailEnv    = "IAM_SUPER_ADMIN_EMAIL"
	superAdminPasswordEnv = "IAM_SUPER_ADMIN_PASSWORD"
	defaultSuperFirstName = "Super"
	defaultSuperLastName  = "Admin"
	superAdminRoleCode    = "super_admin"
)

type SuperAdminSeeder struct {
	transactor         sharedrepo.Transactor
	accountUsecase     usecase.AccountUsecase
	userUsecase        usecase.UserUsecase
	roleRepo           repository.RoleRepository
	roleAssignmentRepo repository.RoleAssignmentRepository
	logger             core.Logger
}

func NewSuperAdminSeeder(
	transactor sharedrepo.Transactor,
	accountUsecase usecase.AccountUsecase,
	userUsecase usecase.UserUsecase,
	roleRepo repository.RoleRepository,
	roleAssignmentRepo repository.RoleAssignmentRepository,
	logger core.Logger,
) *SuperAdminSeeder {
	return &SuperAdminSeeder{
		transactor:         transactor,
		accountUsecase:     accountUsecase,
		userUsecase:        userUsecase,
		roleRepo:           roleRepo,
		roleAssignmentRepo: roleAssignmentRepo,
		logger:             logger,
	}
}

func (s *SuperAdminSeeder) Seed(ctx context.Context) error {
	email, password, ok, err := s.getSeedCredentials()
	if err != nil {
		return err
	}
	if !ok {
		s.logger.Warn("Super admin seed skipped: missing env vars",
			core.String("email_env", superAdminEmailEnv),
			core.String("password_env", superAdminPasswordEnv),
		)
		return nil
	}

	account, err := s.accountUsecase.GetAccountByEmail(ctx, email)
	if err != nil && err != iamerror.ErrAccountNotFound {
		return err
	}

	role, err := s.roleRepo.GetByCode(ctx, superAdminRoleCode)
	if err != nil {
		return err
	}

	if account != nil {
		if err := s.ensureRoleAssignment(ctx, account.ID, role.ID); err != nil {
			return err
		}
		s.logger.Info("Super admin account exists", core.String("accountID", account.ID.String()))
		return nil
	}

	return s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, txErr := s.userUsecase.CreateUser(txCtx, usecase.CreateUserInput{
			FirstName: defaultSuperFirstName,
			LastName:  defaultSuperLastName,
		})
		if txErr != nil {
			return txErr
		}

		passwordHash, txErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if txErr != nil {
			return fmt.Errorf("failed to hash super admin password: %w", txErr)
		}
		passwordHashStr := string(passwordHash)

		account, txErr = s.accountUsecase.CreateAccount(txCtx, usecase.CreateAccountInput{
			UserID:        user.ID,
			Email:         email,
			PasswordHash:  &passwordHashStr,
			EmailVerified: true,
			Status:        entity.AccountStatusActive,
		})
		if txErr != nil {
			return txErr
		}

		if txErr := s.createRoleAssignment(txCtx, account.ID, role.ID); txErr != nil {
			return txErr
		}

		s.logger.Info("Super admin account seeded", core.String("accountID", account.ID.String()))
		return nil
	})
}

func (s *SuperAdminSeeder) getSeedCredentials() (string, string, bool, error) {
	email := strings.TrimSpace(os.Getenv(superAdminEmailEnv))
	password := strings.TrimSpace(os.Getenv(superAdminPasswordEnv))

	if email == "" || password == "" {
		return "", "", false, nil
	}

	normalizedEmail, ok := validation.NormalizeEmail(email)
	if !ok {
		return "", "", false, fmt.Errorf("invalid %s value", superAdminEmailEnv)
	}

	return normalizedEmail, password, true, nil
}

func (s *SuperAdminSeeder) ensureRoleAssignment(ctx context.Context, accountID, roleID uuid.UUID) error {
	exists, err := s.roleAssignmentRepo.ExistsByAccountAndRole(ctx, accountID, roleID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return s.createRoleAssignment(ctx, accountID, roleID)
}

func (s *SuperAdminSeeder) createRoleAssignment(ctx context.Context, accountID, roleID uuid.UUID) error {
	assignment := &entity.RoleAssignment{
		AccountID:  accountID,
		RoleID:     roleID,
		AssignedBy: accountID,
	}

	return s.roleAssignmentRepo.Create(ctx, assignment)
}
