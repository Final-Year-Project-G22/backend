package appusecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/oauth"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type oauthIdentityUsecase struct {
	oauthRepo   repository.OAuthIdentityRepository
	accountRepo repository.AccountRepository
	encryptor   *oauth.TokenEncryptor
	logger      core.Logger
}

func NewOAuthIdentityUsecase(
	oauthRepo repository.OAuthIdentityRepository,
	accountRepo repository.AccountRepository,
	encryptor *oauth.TokenEncryptor,
	logger core.Logger,
) usecase.OAuthIdentityUsecase {
	return &oauthIdentityUsecase{
		oauthRepo:   oauthRepo,
		accountRepo: accountRepo,
		encryptor:   encryptor,
		logger:      logger,
	}
}

func (u *oauthIdentityUsecase) LinkOAuthIdentity(
	ctx context.Context,
	accountID uuid.UUID,
	input usecase.LinkOAuthIdentityInput,
) (*entity.OAuthIdentity, error) {
	existing, err := u.oauthRepo.GetByProviderSubject(ctx, input.Provider, input.ProviderSubject)
	if err == nil && existing != nil {
		return nil, iamerror.ErrOAuthIdentityAlreadyLinked
	}
	if err != nil && err != iamerror.ErrOAuthIdentityNotFound {
		return nil, err
	}

	var encryptedAccess, encryptedRefresh *string

	if input.AccessToken != nil {
		encrypted, err := u.encryptor.Encrypt(*input.AccessToken)
		if err != nil {
			u.logger.Error("Failed to encrypt access token", core.Error(err))
			return nil, errors.InternalError("oauth.errors.tokenEncryptionFailed", err)
		}
		encryptedAccess = &encrypted
	}

	if input.RefreshToken != nil {
		encrypted, err := u.encryptor.Encrypt(*input.RefreshToken)
		if err != nil {
			u.logger.Error("Failed to encrypt refresh token", core.Error(err))
			return nil, errors.InternalError("oauth.errors.tokenEncryptionFailed", err)
		}
		encryptedRefresh = &encrypted
	}

	identity := &entity.OAuthIdentity{
		AccountID:       accountID,
		Provider:        input.Provider,
		ProviderSubject: input.ProviderSubject,
		ProviderEmail:   input.ProviderEmail,
		AccessToken:     encryptedAccess,
		RefreshToken:    encryptedRefresh,
		TokenExpiresAt:  input.TokenExpiresAt,
		LastUsedAt:      timePtr(time.Now()),
	}

	if err := u.oauthRepo.Create(ctx, identity); err != nil {
		return nil, err
	}

	u.logger.Info("OAuth identity linked",
		core.String("identityID", identity.ID.String()),
		core.String("accountID", accountID.String()),
		core.String("provider", input.Provider),
	)

	return identity, nil
}

func (u *oauthIdentityUsecase) GetByProviderSubject(
	ctx context.Context,
	provider, subject string,
) (*entity.OAuthIdentity, error) {
	return u.oauthRepo.GetByProviderSubject(ctx, provider, subject)
}

func (u *oauthIdentityUsecase) ListOAuthIdentities(
	ctx context.Context,
	accountID uuid.UUID,
) ([]*entity.OAuthIdentity, error) {
	return u.oauthRepo.ListByAccountID(ctx, accountID)
}

func (u *oauthIdentityUsecase) UnlinkOAuthIdentity(
	ctx context.Context,
	accountID, identityID uuid.UUID,
) error {
	identity, err := u.oauthRepo.GetByID(ctx, identityID)
	if err != nil {
		return err
	}

	if identity.AccountID != accountID {
		return iamerror.ErrOAuthIdentityNotFound
	}

	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}

	hasPassword := account.PasswordHash != nil && *account.PasswordHash != ""

	if !hasPassword {
		linkedProviders, err := u.oauthRepo.ListByAccountID(ctx, accountID)
		if err != nil {
			return err
		}

		if len(linkedProviders) <= 1 {
			return errors.BadRequestError("oauth.errors.cannotUnlinkLastProvider")
		}
	}

	if err := u.oauthRepo.Delete(ctx, identityID); err != nil {
		return err
	}

	u.logger.Info("OAuth identity unlinked",
		core.String("identityID", identityID.String()),
		core.String("accountID", accountID.String()),
		core.String("provider", identity.Provider),
	)

	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
