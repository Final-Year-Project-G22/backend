package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type AccountEmailOTPRepository interface {
	sharedrepo.GenericRepository[entity.AccountEmailOTP]

	GetLatestByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.AccountEmailOTP, error)
	GetActiveByAccountID(ctx context.Context, accountID uuid.UUID, now time.Time) (*entity.AccountEmailOTP, error)
	GetActiveByAccountIDAndPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) (*entity.AccountEmailOTP, error)
	IncrementAttemptCount(ctx context.Context, otpID uuid.UUID) error
	Consume(ctx context.Context, otpID uuid.UUID, consumedAt time.Time) error
	InvalidateActiveByAccountID(ctx context.Context, accountID uuid.UUID, now time.Time) error
	IncrementResendCountAndUpdateLastSentAt(ctx context.Context, otpID uuid.UUID, now time.Time) error
	InvalidateActiveByAccountIDAndPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) error
	FindActiveByCodeHashAndPurpose(ctx context.Context, codeHash string, purpose string, now time.Time) (*entity.AccountEmailOTP, error)
}
