package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/validation"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/event"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
)

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	ValidateAccessSession(ctx context.Context, sessionID uuid.UUID, checkStatus bool) (ValidatedAccessSessionOutput, error)
	VerifyEmailOTP(ctx context.Context, accountID uuid.UUID, userID uuid.UUID, otp string) error
	ResendEmailOTP(ctx context.Context, accountID uuid.UUID) error
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, accountID uuid.UUID) error
	GetAccountIDBySessionID(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error)
	UpdateUserProfile(ctx context.Context, useId uuid.UUID, input UpdateUserProfileInput) (*UpdateUserProfileOutput, error)
	UpdateAccountPassword(ctx context.Context, userId uuid.UUID, input UpdateAccountPasswordInput) error
	GetCurrentUser(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (*GetCurrentUserOutput, error)
}

type RegisterInput struct {
	Email     string
	Username  *string
	Password  string
	FirstName string
	LastName  string
	UserAgent *string
	IPAddress *string
}

type LoginInput struct {
	Identifier string
	Password   string
	UserAgent  *string
	IPAddress  *string
}
type UpdateUserProfileInput struct {
	FirstName string
	LastName  string
	Bio       *string
}
type UpdateUserProfileOutput struct {
	FirstName string
	LastName  string
	Bio       string
}
type UpdateAccountPasswordInput struct {
	ExistingPassword string
	NewPassword      string
	ConfirmPassword  string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         *entity.User
	Account      *entity.Account
}

type ValidatedAccessSessionOutput struct {
	Email     string
	AccountID uuid.UUID
	UserID    uuid.UUID
}
type GetCurrentUserOutput struct {
	User        *entity.User
	Account     *entity.Account
	Roles       []*entity.Role
	Permissions []*entity.Permission
}

type authService struct {
	transactor            sharedrepo.Transactor
	userUsecase           usecase.UserUsecase
	accountUsecase        usecase.AccountUsecase
	otpUsecase            usecase.AccountEmailOTPUsecase
	sessionUsecase        usecase.SessionUsecase
	roleAssignmentUsecase usecase.RoleAssignmentUsecase
	sessionRepo           repository.SessionRepository
	tokenService          token.TokenService
	logger                core.Logger
	cfg                   *core.Config
	notifOutboxRepo       notifrepo.NotificationOutboxRepository
}

const (
	otpTTL           = 3 * time.Minute
	otpResendCoolOff = 60 * time.Second
	otpMaxAttempts   = 5
	otpMaxResends    = 5
)

func NewAuthService(
	transactor sharedrepo.Transactor,
	userUsecase usecase.UserUsecase,
	accountUsecase usecase.AccountUsecase,
	otpUsecase usecase.AccountEmailOTPUsecase,
	sessionUsecase usecase.SessionUsecase,
	roleAssignmentUsecase usecase.RoleAssignmentUsecase,
	sessionRepo repository.SessionRepository,
	tokenService token.TokenService,
	logger core.Logger,
	cfg *core.Config,
	notifOutboxRepo notifrepo.NotificationOutboxRepository,
) AuthService {
	return &authService{
		transactor:            transactor,
		userUsecase:           userUsecase,
		accountUsecase:        accountUsecase,
		otpUsecase:            otpUsecase,
		sessionUsecase:        sessionUsecase,
		roleAssignmentUsecase: roleAssignmentUsecase,
		sessionRepo:           sessionRepo,
		tokenService:          tokenService,
		logger:                logger,
		cfg:                   cfg,
		notifOutboxRepo:       notifOutboxRepo,
	}
}

// Register creates a new user, account, and session atomically.
// Returns tokens immediately (user is logged in after registration).
func (s *authService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	email, ok := validation.NormalizeEmail(input.Email)
	if !ok {
		return nil, errors.BadRequestError("iam.errors.invalidEmailFormat")
	}

	if input.Username != nil && strings.TrimSpace(*input.Username) != "" {
		if _, valid := validation.NormalizeUsername(*input.Username); !valid {
			return nil, errors.BadRequestError("iam.errors.invalidUsernameFormat")
		}
	}

	// Check if account already exists
	existingAccount, err := s.accountUsecase.GetAccountByEmail(ctx, email)
	if err == nil && existingAccount != nil {
		return nil, errors.ConflictError("iam.errors.emailAlreadyExists")
	}
	if err != nil && err != iamerror.ErrAccountNotFound {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", core.Error(err))
		return nil, errors.InternalError("iam.errors.passwordHashFailed", err)
	}

	// Generate refresh token
	rawRefreshToken, refreshTokenHash, err := s.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	var user *entity.User
	var account *entity.Account
	var session *entity.Session
	var otpCode string

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		var txErr error

		otpCode, txErr = generateOTPCode()
		if txErr != nil {
			return errors.InternalError("iam.errors.failedToGenerateOtp", txErr)
		}

		// Create user
		user, txErr = s.userUsecase.CreateUser(txCtx, usecase.CreateUserInput{
			FirstName: input.FirstName,
			LastName:  input.LastName,
		})
		if txErr != nil {
			return txErr
		}

		// Create account
		passwordHashStr := string(passwordHash)
		account, txErr = s.accountUsecase.CreateAccount(txCtx, usecase.CreateAccountInput{
			UserID:        user.ID,
			Email:         email,
			Username:      input.Username,
			PasswordHash:  &passwordHashStr,
			EmailVerified: false,
			Status:        entity.AccountStatusPendingVerification,
		})
		if txErr != nil {
			return txErr
		}

		// Create session
		session, txErr = s.sessionUsecase.CreateSession(txCtx, account.ID, usecase.CreateSessionInput{
			RefreshTokenHash: refreshTokenHash,
			UserAgent:        input.UserAgent,
			IPAddress:        input.IPAddress,
			ExpiresAt:        time.Now().Add(s.tokenService.GetRefreshTokenTTL()),
		})
		if txErr != nil {
			return txErr
		}

		now := time.Now()
		otpRecord, txErr := s.otpUsecase.CreateOTP(txCtx, account.ID, hashOTPCode(otpCode), now.Add(otpTTL), 0, now, string(entity.EmailOTPPurposeVerification))
		if txErr != nil {
			return txErr
		}

		if txErr := s.writeOTPNotificationOutbox(txCtx, account.ID, account.Email, user, otpCode, otpRecord.ID); txErr != nil {
			s.logger.Error("Failed to write OTP notification outbox row", core.Error(txErr))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Generate access token (outside transaction - read-only operation)
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, token.AccessTokenClaims{
		SessionID: session.ID,
		Email:     account.Email,
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("User registered successfully",
		core.String("userID", user.ID.String()),
		core.String("accountID", account.ID.String()),
	)

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    time.Now().Add(s.tokenService.GetAccessTokenTTL()),
		User:         user,
		Account:      account,
	}, nil
}

// Login authenticates a user and creates a new session.
func (s *authService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	_, identifier, ok := validation.NormalizeIdentifier(input.Identifier)
	if !ok {
		return nil, errors.BadRequestError("iam.errors.invalidIdentifier")
	}

	// Get account by email or username
	account, err := s.accountUsecase.GetAccountByIdentifier(ctx, identifier)
	if err != nil {
		if err == iamerror.ErrAccountNotFound {
			return nil, errors.UnauthorizedError("iam.errors.invalidCredentials")
		}
		return nil, err
	}

	// Verify password
	if account.PasswordHash == nil {
		return nil, errors.UnauthorizedError("iam.errors.invalidCredentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.UnauthorizedError("iam.errors.invalidCredentials")
	}

	// Check account status
	if err := s.ensureAccountCanAuthenticate(account); err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	rawRefreshToken, refreshTokenHash, err := s.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	// Create session
	session, err := s.sessionUsecase.CreateSession(ctx, account.ID, usecase.CreateSessionInput{
		RefreshTokenHash: refreshTokenHash,
		UserAgent:        input.UserAgent,
		IPAddress:        input.IPAddress,
		ExpiresAt:        time.Now().Add(s.tokenService.GetRefreshTokenTTL()),
	})
	if err != nil {
		return nil, err
	}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, token.AccessTokenClaims{
		SessionID: session.ID,
		Email:     account.Email,
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("User logged in successfully",
		core.String("accountID", account.ID.String()),
	)

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    time.Now().Add(s.tokenService.GetAccessTokenTTL()),
		User:         user,
		Account:      account,
	}, nil
}
func (s *authService) UpdateUserProfile(ctx context.Context, userId uuid.UUID, input UpdateUserProfileInput) (*UpdateUserProfileOutput, error) {
	user, err := s.userUsecase.UpdateUser(ctx, userId, usecase.UpdateUserInput{
		FirstName: &input.FirstName,
		LastName:  &input.LastName,
		Bio:       input.Bio,
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("User Profile updated Successfully",
		core.String("userID", user.ID.String()),
	)
	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}
	return &UpdateUserProfileOutput{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Bio:       bio,
	}, nil
}

func (s *authService) UpdateAccountPassword(ctx context.Context, accountID uuid.UUID, input UpdateAccountPasswordInput) error {

	account, err := s.accountUsecase.GetAccount(ctx, accountID)

	if err != nil {
		return err
	}

	if account.PasswordHash == nil {
		return errors.UnauthorizedError("iam.errors.invalidPassword")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(input.ExistingPassword)); err != nil {
		return errors.UnauthorizedError("iam.errors.invalidPassword")
	}

	if input.NewPassword != input.ConfirmPassword {
		return errors.BadRequestError("iam.errors.passwordMismatch")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", core.Error(err))
		return errors.InternalError("iam.errors.passwordHashFailed", err)
	}

	user, err := s.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return err
	}

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if txErr := s.accountUsecase.UpdateAccountPassword(txCtx, account.ID, usecase.UpdateAccountPasswordInput{
			NewHashedPassword: string(hashedPassword),
		}); txErr != nil {
			return txErr
		}

		if txErr := s.writeAccountAlertOutbox(txCtx, account, user, "password_changed", "Password Changed", "Your account password has been changed successfully.", "/account/security"); txErr != nil {
			s.logger.Error("Failed to write account alert notification outbox row", core.Error(txErr))
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info("Account password updated successfully", core.String("accountID", account.ID.String()))
	return nil

}

// Refresh rotates the refresh token and issues new tokens.
// Old session is revoked, new session is created.
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Hash the provided refresh token to look up the session
	tokenHash := s.tokenService.HashRefreshToken(refreshToken)

	// Get session by refresh token hash
	session, err := s.sessionUsecase.GetSessionByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		if err == iamerror.ErrSessionNotFound {
			return nil, errors.UnauthorizedError("iam.errors.invalidRefreshToken")
		}
		return nil, err
	}

	// Get account
	account, err := s.accountUsecase.GetAccount(ctx, session.AccountID)
	if err != nil {
		return nil, err
	}

	// Check account status
	if err := s.ensureAccountCanAuthenticate(account); err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	rawRefreshToken, refreshTokenHash, err := s.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	var newSession *entity.Session

	// Rotate: revoke old session and create new one atomically
	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Revoke old session
		if txErr := s.sessionUsecase.RevokeSession(txCtx, session.ID); txErr != nil {
			return txErr
		}

		// Create new session
		var txErr error
		newSession, txErr = s.sessionUsecase.CreateSession(txCtx, account.ID, usecase.CreateSessionInput{
			RefreshTokenHash: refreshTokenHash,
			UserAgent:        session.UserAgent,
			IPAddress:        session.IPAddress,
			ExpiresAt:        time.Now().Add(s.tokenService.GetRefreshTokenTTL()),
		})
		return txErr
	})

	if err != nil {
		return nil, err
	}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, token.AccessTokenClaims{
		SessionID: newSession.ID,
		Email:     account.Email,
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Session refreshed successfully",
		core.String("oldSessionID", session.ID.String()),
		core.String("newSessionID", newSession.ID.String()),
		core.String("accountID", account.ID.String()),
	)

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    time.Now().Add(s.tokenService.GetAccessTokenTTL()),
		User:         user,
		Account:      account,
	}, nil
}

func (s *authService) ValidateAccessSession(ctx context.Context, sessionID uuid.UUID, checkStatus bool) (ValidatedAccessSessionOutput, error) {
	session, err := s.sessionRepo.GetActiveByID(ctx, sessionID)
	if err != nil {
		if err == iamerror.ErrSessionNotFound {
			return ValidatedAccessSessionOutput{}, errors.UnauthorizedError("iam.errors.sessionNotFound")
		}
		return ValidatedAccessSessionOutput{}, err
	}

	account, err := s.accountUsecase.GetAccount(ctx, session.AccountID)
	if err != nil {
		return ValidatedAccessSessionOutput{}, err
	}

	if checkStatus {
		if err := s.ensureAccountCanAuthenticate(account); err != nil {
			return ValidatedAccessSessionOutput{}, err
		}
	}

	return ValidatedAccessSessionOutput{
		Email:     account.Email,
		AccountID: account.ID,
		UserID:    account.UserID,
	}, nil
}

func (s *authService) VerifyEmailOTP(ctx context.Context, accountID uuid.UUID, userID uuid.UUID, otp string) error {
	account, err := s.accountUsecase.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}
	user, err := s.userUsecase.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if account.EmailVerified {
		return nil
	}

	now := time.Now()
	activeOTP, err := s.otpUsecase.GetActiveOTP(ctx, accountID, now)
	if err != nil {
		if err == iamerror.ErrEmailOTPNotFound {
			return errors.BadRequestError("iam.errors.invalidOtp")
		}
		return err
	}

	if activeOTP.AttemptCount >= otpMaxAttempts {
		return errors.BadRequestError("iam.errors.otpAttemptsExceeded")
	}

	inputHash := hashOTPCode(otp)
	if subtle.ConstantTimeCompare([]byte(inputHash), []byte(activeOTP.CodeHash)) != 1 {
		if err := s.otpUsecase.IncrementAttemptCount(ctx, activeOTP.ID); err != nil {
			return err
		}
		return errors.BadRequestError("iam.errors.invalidOtp")
	}

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if txErr := s.otpUsecase.ConsumeOTP(txCtx, activeOTP.ID, now); txErr != nil {
			return txErr
		}

		if txErr := s.accountUsecase.MarkEmailVerifiedAndActivate(txCtx, accountID); txErr != nil {
			return txErr
		}

		if txErr := s.writeWelcomeNotificationOutbox(txCtx, account, user); txErr != nil {
			s.logger.Error("Failed to write welcome notification outbox row", core.Error(txErr))
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) ResendEmailOTP(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.accountUsecase.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	if account.EmailVerified {
		return errors.BadRequestError("iam.errors.emailAlreadyVerified")
	}

	user, err := s.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return err
	}

	now := time.Now()
	latestOTP, err := s.otpUsecase.GetLatestOTP(ctx, accountID)
	if err != nil && err != iamerror.ErrEmailOTPNotFound {
		return err
	}

	if err == nil {
		if now.Sub(latestOTP.LastSentAt) < otpResendCoolOff {
			return errors.BadRequestError("iam.errors.otpResendTooSoon")
		}
		if latestOTP.ResendCount >= otpMaxResends {
			return errors.BadRequestError("iam.errors.otpResendLimitExceeded")
		}
	}

	otpCode, err := generateOTPCode()
	if err != nil {
		return errors.InternalError("iam.errors.failedToGenerateOtp", err)
	}

	resendCount := 1
	if latestOTP != nil {
		resendCount = latestOTP.ResendCount + 1
	}

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if txErr := s.otpUsecase.InvalidateActiveOTP(txCtx, accountID, now); txErr != nil {
			return txErr
		}

		otpRecord, txErr := s.otpUsecase.CreateOTP(txCtx, accountID, hashOTPCode(otpCode), now.Add(otpTTL), resendCount, now, string(entity.EmailOTPPurposeVerification))
		if txErr != nil {
			return txErr
		}

		if txErr := s.writeOTPNotificationOutbox(txCtx, accountID, account.Email, user, otpCode, otpRecord.ID); txErr != nil {
			s.logger.Error("Failed to write OTP notification outbox row", core.Error(txErr))
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) ensureAccountCanAuthenticate(account *entity.Account) error {
	switch account.Status {
	case entity.AccountStatusActive:
		return nil
	case entity.AccountStatusPendingVerification:
		return errors.ForbiddenError("iam.errors.accountStatusPendingVerification")
	case entity.AccountStatusLocked:
		return errors.ForbiddenError("iam.errors.accountStatusLocked")
	case entity.AccountStatusSuspended:
		return errors.ForbiddenError("iam.errors.accountStatusSuspended")
	case entity.AccountStatusDisabled:
		return errors.ForbiddenError("iam.errors.accountStatusDisabled")
	default:
		return errors.ForbiddenError("iam.errors.forbidden")
	}
}

// Logout revokes a specific session.
func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionUsecase.RevokeSession(ctx, sessionID); err != nil {
		return err
	}

	s.logger.Info("Session logged out", core.String("sessionID", sessionID.String()))
	return nil
}

// LogoutAll revokes all sessions for an account.
func (s *authService) LogoutAll(ctx context.Context, accountID uuid.UUID) error {
	if err := s.sessionUsecase.RevokeAllSessions(ctx, accountID); err != nil {
		return err
	}

	s.logger.Info("All sessions logged out", core.String("accountID", accountID.String()))
	return nil
}

func (s *authService) GetAccountIDBySessionID(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if err == iamerror.ErrSessionNotFound {
			return uuid.Nil, errors.UnauthorizedError("iam.errors.sessionNotFound")
		}
		return uuid.Nil, err
	}
	return session.AccountID, nil
}

func (s *authService) GetCurrentUser(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (*GetCurrentUserOutput, error) {
	user, err := s.userUsecase.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	account, err := s.accountUsecase.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	roles, err := s.roleAssignmentUsecase.ListAccountRoles(ctx, accountID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.roleAssignmentUsecase.GetEffectivePermissions(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &GetCurrentUserOutput{
		User:        user,
		Account:     account,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func (s *authService) writeWelcomeNotificationOutbox(ctx context.Context, account *entity.Account, user *entity.User) error {
	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        event.AccountRegistered,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "iam",
		AccountID:        account.ID,
		NotificationType: string(notifentity.NotificationTypeWelcomeMessage),
		ChannelPolicy:    notificationevent.ChannelPolicyAllEnabled,
		Variables: map[string]string{
			"platformName":      s.cfg.App.Name,
			"accountName":       user.FirstName,
			"gettingStartedUrl": "/guides",
		},
		Metadata: notificationevent.Metadata{
			IdempotencyKey: "welcome:" + account.ID.String(),
			Locale:         nil,
		},
	}
	return s.writeEnvelopeToOutbox(ctx, &env)
}

func (s *authService) writeOTPNotificationOutbox(ctx context.Context, accountID uuid.UUID, email string, user *entity.User, otpCode string, otpRecordID uuid.UUID) error {
	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        event.UserEmailOTPRequested,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "iam",
		AccountID:        accountID,
		NotificationType: string(notifentity.NotificationTypeAccountVerification),
		ChannelPolicy:    notificationevent.ChannelPolicySingle,
		Channel:          strPtr(string(notifentity.ChannelEmail)),
		Variables: map[string]string{
			"platformName":        s.cfg.App.Name,
			"verificationMessage": fmt.Sprintf("Your verification code is %s. Please enter this code to verify your email.", otpCode),
			"verificationUrl":     "https://app.wegotcha.com/verify-email?code=" + otpCode,
			"expiryMinutes":       fmt.Sprintf("%d", int(otpTTL/time.Minute)),
		},
		Metadata: notificationevent.Metadata{
			IdempotencyKey: "verify-email:" + accountID.String() + ":" + otpRecordID.String(),
			Locale:         nil,
		},
	}
	return s.writeEnvelopeToOutbox(ctx, &env)
}

func (s *authService) writeAccountAlertOutbox(ctx context.Context, account *entity.Account, user *entity.User, alertCode, alertTitle, alertMessage, securityUrl string) error {
	notificationType := string(notifentity.NotificationTypeAccountAlertCritical)
	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        event.AccountAlert,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "iam",
		AccountID:        account.ID,
		NotificationType: notificationType,
		ChannelPolicy:    notificationevent.ChannelPolicyAllEnabled,
		Variables: map[string]string{
			"alertTitle":   alertTitle,
			"alertMessage": alertMessage,
			"alertCode":    alertCode,
			"securityUrl":  securityUrl,
		},
		Metadata: notificationevent.Metadata{
			IdempotencyKey: "account-alert:" + account.ID.String() + ":" + alertCode + ":" + uuid.New().String(),
			Locale:         nil,
		},
	}
	return s.writeEnvelopeToOutbox(ctx, &env)
}

func (s *authService) writeEnvelopeToOutbox(ctx context.Context, envelope *notificationevent.Envelope) error {
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	var payload datatypes.JSONMap
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to convert envelope to JSONMap: %w", err)
	}

	outbox := &notifentity.NotificationOutbox{
		EventType:      envelope.EventType,
		SchemaVersion:  envelope.SchemaVersion,
		SourceModule:   envelope.SourceModule,
		AccountID:      envelope.AccountID,
		IdempotencyKey: envelope.Metadata.IdempotencyKey,
		Payload:        payload,
		Status:         notifentity.NotificationOutboxStatusPending,
		AttemptCount:   0,
	}

	return s.notifOutboxRepo.Create(ctx, outbox)
}

func strPtr(s string) *string {
	return &s
}

func generateOTPCode() (string, error) {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", randomInt.Int64()), nil
}

func hashOTPCode(otpCode string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(otpCode)))
	return hex.EncodeToString(hash[:])
}
