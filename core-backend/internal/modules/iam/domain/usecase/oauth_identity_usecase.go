package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type OAuthIdentityUsecase interface {
	LinkOAuthIdentity(ctx context.Context, accountID uuid.UUID, input LinkOAuthIdentityInput) (*entity.OAuthIdentity, error)
	GetByProviderSubject(ctx context.Context, provider, subject string) (*entity.OAuthIdentity, error)
	ListOAuthIdentities(ctx context.Context, accountID uuid.UUID) ([]*entity.OAuthIdentity, error)
	UnlinkOAuthIdentity(ctx context.Context, accountID, identityID uuid.UUID) error
}

type LinkOAuthIdentityInput struct {
	Provider        string
	ProviderSubject string
	ProviderEmail   *string
	AccessToken     *string
	RefreshToken    *string
	TokenExpiresAt  *time.Time
}
