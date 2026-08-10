package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
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
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
)

const adminPasswordLength = 12

type AdminService interface {
	RegisterAdmin(ctx context.Context, input RegisterAdminInput) (*RegisterAdminOutput, error)
	UpdateAdminRoles(ctx context.Context, input UpdateAdminRolesInput) error
	ListAdmins(ctx context.Context, input ListAdminsInput) (*ListAdminsOutput, error)
	UpdateAdminStatus(ctx context.Context, input UpdateAdminStatusInput) error
	ResetAdminPassword(ctx context.Context, input ResetAdminPasswordInput) error
	CompletePasswordReset(ctx context.Context, token string, newPassword string) error
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

type ListAdminsInput struct {
	Search   string
	Status   string
	RoleID   string
	Page     int
	PageSize int
}

type ListAdminsOutput struct {
	Admins     []AdminAccountInfo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type AdminAccountInfo struct {
	ID        uuid.UUID
	Email     string
	Username  *string
	Status    string
	FirstName string
	LastName  string
	Roles     []RoleInfo
	CreatedAt string
	LastLogin *string
}

type RoleInfo struct {
	ID   uuid.UUID
	Code string
	Name string
}

type UpdateAdminStatusInput struct {
	AccountID uuid.UUID
	Status    entity.AccountStatus
}

type ResetAdminPasswordInput struct {
	AccountID   uuid.UUID
	TriggeredBy uuid.UUID
}

type adminService struct {
	transactor         sharedrepo.Transactor
	userUsecase        usecase.UserUsecase
	accountUsecase     usecase.AccountUsecase
	roleRepo           repository.RoleRepository
	roleAssignmentRepo repository.RoleAssignmentRepository
	otpUsecase         usecase.AccountEmailOTPUsecase
	messageBus         rabbitmq.Bus
	logger             core.Logger
	cfg                *core.Config
	notifOutboxRepo    notifrepo.NotificationOutboxRepository
}

func NewAdminService(
	transactor sharedrepo.Transactor,
	userUsecase usecase.UserUsecase,
	accountUsecase usecase.AccountUsecase,
	roleRepo repository.RoleRepository,
	roleAssignmentRepo repository.RoleAssignmentRepository,
	otpUsecase usecase.AccountEmailOTPUsecase,
	messageBus rabbitmq.Bus,
	logger core.Logger,
	cfg *core.Config,
	notifOutboxRepo notifrepo.NotificationOutboxRepository,
) AdminService {
	return &adminService{
		transactor:         transactor,
		userUsecase:        userUsecase,
		accountUsecase:     accountUsecase,
		roleRepo:           roleRepo,
		roleAssignmentRepo: roleAssignmentRepo,
		otpUsecase:         otpUsecase,
		messageBus:         messageBus,
		logger:             logger,
		cfg:                cfg,
		notifOutboxRepo:    notifOutboxRepo,
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

		if txErr := s.writeAdminCreatedOutbox(txCtx, account, user, password); txErr != nil {
			s.logger.Error("Failed to write admin credentials notification outbox row", core.Error(txErr))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

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
			if err := s.roleAssignmentRepo.Revoke(txCtx, assignment.ID, time.Now(), permissions.StringPtr("roles updated")); err != nil {
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

func (s *adminService) ListAdmins(ctx context.Context, input ListAdminsInput) (*ListAdminsOutput, error) {
	// An account is an admin if any of its roles grants any IAM-admin capability
	// (list/read/create/roles.update/status.update/reset_password) or role
	// management. Admin hub registration offers every role, so admins registered
	// with content/moderation roles (e.g. ai_content_manager carries
	// iam.admin.read) must still appear in the hub's admin list.
	permissionCodes := []string{
		permissions.AdminList,
		permissions.AdminRead,
		permissions.AdminCreate,
		permissions.AdminRolesUpdate,
		permissions.AdminResetPassword,
		permissions.AdminStatusUpdate,
		permissions.RoleRead,
	}
	queryOpts := map[string]interface{}{
		"search":   input.Search,
		"status":   input.Status,
		"roleId":   input.RoleID,
		"page":     input.Page,
		"pageSize": input.PageSize,
	}

	accounts, total, err := s.accountUsecase.ListAdmins(ctx, permissionCodes, queryOpts)
	if err != nil {
		return nil, err
	}

	admins := make([]AdminAccountInfo, 0, len(accounts))
	for _, account := range accounts {
		user, _ := s.userUsecase.GetUser(ctx, account.UserID)
		assignments, _ := s.roleAssignmentRepo.ListByAccountID(ctx, account.ID)
		roles := make([]RoleInfo, 0)
		for _, a := range assignments {
			if a.RevokedAt != nil {
				continue
			}
			role, _ := s.roleRepo.GetByID(ctx, a.RoleID)
			if role != nil {
				roles = append(roles, RoleInfo{
					ID:   role.ID,
					Code: role.Code,
					Name: role.Name,
				})
			}
		}
		firstName, lastName := "", ""
		if user != nil {
			firstName = user.FirstName
			lastName = user.LastName
		}
		var lastLogin *string
		if account.LastLoginAt != nil {
			t := account.LastLoginAt.Format(time.RFC3339)
			lastLogin = &t
		}
		createdAt := account.CreatedAt.Format(time.RFC3339)
		admins = append(admins, AdminAccountInfo{
			ID:        account.ID,
			Email:     account.Email,
			Username:  account.Username,
			Status:    string(account.Status),
			FirstName: firstName,
			LastName:  lastName,
			Roles:     roles,
			CreatedAt: createdAt,
			LastLogin: lastLogin,
		})
	}

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ListAdminsOutput{
		Admins:     admins,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *adminService) UpdateAdminStatus(ctx context.Context, input UpdateAdminStatusInput) error {
	validTransitions := map[entity.AccountStatus][]entity.AccountStatus{
		entity.AccountStatusActive:    {entity.AccountStatusLocked, entity.AccountStatusSuspended},
		entity.AccountStatusLocked:    {entity.AccountStatusActive},
		entity.AccountStatusSuspended: {entity.AccountStatusActive},
	}
	account, err := s.accountUsecase.GetAccount(ctx, input.AccountID)
	if err != nil {
		return err
	}
	allowed, ok := validTransitions[account.Status]
	if !ok {
		return errors.BadRequestError("iam.errors.invalidStatusTransition")
	}
	allowedMap := make(map[entity.AccountStatus]bool)
	for _, s := range allowed {
		allowedMap[s] = true
	}
	if !allowedMap[input.Status] {
		return errors.BadRequestError("iam.errors.invalidStatusTransition")
	}
	return s.accountUsecase.ChangeAccountStatus(ctx, input.AccountID, input.Status)
}

func (s *adminService) ResetAdminPassword(ctx context.Context, input ResetAdminPasswordInput) error {
	account, err := s.accountUsecase.GetAccount(ctx, input.AccountID)
	if err != nil {
		return err
	}
	user, err := s.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return err
	}
	otpCode, err := generateAdminPassword(8)
	if err != nil {
		return errors.InternalError("iam.errors.failedToGenerateOtp", err)
	}
	now := time.Now()
	resetTTL := 5 * time.Minute
	if _, err := s.otpUsecase.CreateOTP(ctx, account.ID, hashOTPCode(otpCode), now.Add(resetTTL), 0, now, string(entity.EmailOTPPurposePasswordReset)); err != nil {
		return err
	}
	s.publishPasswordReset(ctx, account, user, otpCode)
	return nil
}

func (s *adminService) CompletePasswordReset(ctx context.Context, token string, newPassword string) error {
	accounts, _, err := s.accountUsecase.ListAdmins(ctx, []string{}, map[string]interface{}{"page": 1, "pageSize": 1000})
	if err != nil {
		return err
	}
	var found *entity.Account
	var activeOTP *entity.AccountEmailOTP
	now := time.Now()
	for _, account := range accounts {
		_, err := s.userUsecase.GetUser(ctx, account.UserID)
		if err != nil {
			continue
		}
		otp, err := s.otpUsecase.GetActiveOTPByPurpose(ctx, account.ID, string(entity.EmailOTPPurposePasswordReset), now)
		if err != nil || otp == nil {
			continue
		}
		inputHash := hashOTPCode(token)
		if subtle.ConstantTimeCompare([]byte(inputHash), []byte(otp.CodeHash)) == 1 {
			found = account
			activeOTP = otp
			break
		}
	}
	if found == nil {
		return errors.BadRequestError("iam.errors.invalidOtp")
	}
	if activeOTP.AttemptCount >= 5 {
		return errors.BadRequestError("iam.errors.otpAttemptsExceeded")
	}
	inputHash := hashOTPCode(token)
	if subtle.ConstantTimeCompare([]byte(inputHash), []byte(activeOTP.CodeHash)) != 1 {
		_ = s.otpUsecase.IncrementAttemptCount(ctx, activeOTP.ID)
		return errors.BadRequestError("iam.errors.invalidOtp")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalError("iam.errors.passwordHashFailed", err)
	}
	if err := s.accountUsecase.UpdateAccountPassword(ctx, found.ID, usecase.UpdateAccountPasswordInput{NewHashedPassword: string(hashed)}); err != nil {
		return err
	}
	_ = s.otpUsecase.ConsumeOTP(ctx, activeOTP.ID, now)
	return nil
}

func (s *adminService) publishPasswordReset(ctx context.Context, account *entity.Account, user *entity.User, otpCode string) {
	err := s.messageBus.Publish(ctx, event.UserEmailOTPRequested, map[string]interface{}{
		"accountId": account.ID.String(),
		"email":     account.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"otpCode":   otpCode,
		"purpose":   "password_reset",
		"locale":    i18n.LocaleFromContext(ctx),
	})
	if err != nil {
		s.logger.Error("Failed to publish password reset event", core.Error(err))
	}
}

func (s *adminService) writeAdminCreatedOutbox(ctx context.Context, account *entity.Account, user *entity.User, password string) error {
	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        event.AdminCreated,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "iam",
		AccountID:        account.ID,
		NotificationType: string(notifentity.NotificationTypeAdminProvisioned),
		ChannelPolicy:    notificationevent.ChannelPolicySingle,
		Channel:          strPtr(string(notifentity.ChannelEmail)),
		Variables: map[string]string{
			"platformName": s.cfg.App.Name,
			"accountName":  user.FirstName,
			"email":        account.Email,
			"password":     password,
		},
		Metadata: notificationevent.Metadata{
			IdempotencyKey: "admin-created:" + account.ID.String() + ":" + uuid.New().String(),
			Locale:         nil,
		},
	}
	return s.writeEnvelopeToOutbox(ctx, &env)
}

func (s *adminService) writeEnvelopeToOutbox(ctx context.Context, envelope *notificationevent.Envelope) error {
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
