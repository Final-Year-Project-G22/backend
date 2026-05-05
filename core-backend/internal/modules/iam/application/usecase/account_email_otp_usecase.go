package appusecase

import (
	"context"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/google/uuid"
)

type accountEmailOTPUsecase struct {
	otpRepo repository.AccountEmailOTPRepository
	logger  core.Logger
}

func NewAccountEmailOTPUsecase(
	otpRepo repository.AccountEmailOTPRepository,
	logger core.Logger,
) usecase.AccountEmailOTPUsecase {
	return &accountEmailOTPUsecase{
		otpRepo: otpRepo,
		logger:  logger,
	}
}

func (u *accountEmailOTPUsecase) CreateOTP(ctx context.Context, accountID uuid.UUID, codeHash string, expiresAt time.Time, resendCount int, lastSentAt time.Time, purpose string) (*entity.AccountEmailOTP, error) {
	otp := &entity.AccountEmailOTP{
		AccountID:    accountID,
		CodeHash:     strings.TrimSpace(codeHash),
		ExpiresAt:    expiresAt,
		ResendCount:  resendCount,
		LastSentAt:   lastSentAt,
		AttemptCount: 0,
		Purpose:      purpose,
	}

	if err := u.otpRepo.Create(ctx, otp); err != nil {
		return nil, err
	}

	u.logger.Info("Account email OTP created",
		core.String("otpID", otp.ID.String()),
		core.String("accountID", accountID.String()),
	)

	return otp, nil
}

func (u *accountEmailOTPUsecase) GetLatestOTP(ctx context.Context, accountID uuid.UUID) (*entity.AccountEmailOTP, error) {
	return u.otpRepo.GetLatestByAccountID(ctx, accountID)
}

func (u *accountEmailOTPUsecase) GetActiveOTP(ctx context.Context, accountID uuid.UUID, now time.Time) (*entity.AccountEmailOTP, error) {
	return u.otpRepo.GetActiveByAccountID(ctx, accountID, now)
}

func (u *accountEmailOTPUsecase) GetActiveOTPByPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) (*entity.AccountEmailOTP, error) {
	return u.otpRepo.GetActiveByAccountIDAndPurpose(ctx, accountID, purpose, now)
}

func (u *accountEmailOTPUsecase) IncrementAttemptCount(ctx context.Context, otpID uuid.UUID) error {
	return u.otpRepo.IncrementAttemptCount(ctx, otpID)
}

func (u *accountEmailOTPUsecase) ConsumeOTP(ctx context.Context, otpID uuid.UUID, consumedAt time.Time) error {
	return u.otpRepo.Consume(ctx, otpID, consumedAt)
}

func (u *accountEmailOTPUsecase) InvalidateActiveOTP(ctx context.Context, accountID uuid.UUID, now time.Time) error {
	return u.otpRepo.InvalidateActiveByAccountID(ctx, accountID, now)
}

func (u *accountEmailOTPUsecase) InvalidateActiveOTPByPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) error {
	return u.otpRepo.InvalidateActiveByAccountIDAndPurpose(ctx, accountID, purpose, now)
}

func (u *accountEmailOTPUsecase) IncrementResendCountAndUpdateLastSentAt(ctx context.Context, otpID uuid.UUID, now time.Time) error {
	return u.otpRepo.IncrementResendCountAndUpdateLastSentAt(ctx, otpID, now)
}

func (u *accountEmailOTPUsecase) FindActiveByCodeHashAndPurpose(ctx context.Context, codeHash string, purpose string, now time.Time) (*entity.AccountEmailOTP, error) {
	return u.otpRepo.FindActiveByCodeHashAndPurpose(ctx, codeHash, purpose, now)
}
