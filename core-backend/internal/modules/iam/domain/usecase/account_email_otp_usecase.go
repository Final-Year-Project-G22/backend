package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type AccountEmailOTPUsecase interface {
	CreateOTP(ctx context.Context, accountID uuid.UUID, codeHash string, expiresAt time.Time, resendCount int, lastSentAt time.Time, purpose string) (*entity.AccountEmailOTP, error)
	GetLatestOTP(ctx context.Context, accountID uuid.UUID) (*entity.AccountEmailOTP, error)
	GetActiveOTP(ctx context.Context, accountID uuid.UUID, now time.Time) (*entity.AccountEmailOTP, error)
	GetActiveOTPByPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) (*entity.AccountEmailOTP, error)
	IncrementAttemptCount(ctx context.Context, otpID uuid.UUID) error
	ConsumeOTP(ctx context.Context, otpID uuid.UUID, consumedAt time.Time) error
	InvalidateActiveOTP(ctx context.Context, accountID uuid.UUID, now time.Time) error
	InvalidateActiveOTPByPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) error
	IncrementResendCountAndUpdateLastSentAt(ctx context.Context, otpID uuid.UUID, now time.Time) error
	FindActiveByCodeHashAndPurpose(ctx context.Context, codeHash string, purpose string, now time.Time) (*entity.AccountEmailOTP, error)
}
