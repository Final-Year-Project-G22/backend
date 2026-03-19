package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	iamdomain "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/oauth"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
)

type oauthService struct {
	providerRegistry *ProviderRegistry
	stateManager     *StateManager

	authService    service.AuthService
	userUsecase    usecase.UserUsecase
	accountUsecase usecase.AccountUsecase
	sessionUsecase usecase.SessionUsecase
	oauthUsecase   usecase.OAuthIdentityUsecase
	tokenService   token.TokenService
	logger         core.Logger
}

func NewOAuthService(
	providerRegistry *ProviderRegistry,
	stateManager *StateManager,
	authService service.AuthService,
	userUsecase usecase.UserUsecase,
	accountUsecase usecase.AccountUsecase,
	sessionUsecase usecase.SessionUsecase,
	oauthUsecase usecase.OAuthIdentityUsecase,
	tokenService token.TokenService,
	logger core.Logger,
) iamdomain.OAuthService {
	return &oauthService{
		providerRegistry: providerRegistry,
		stateManager:     stateManager,
		authService:      authService,
		userUsecase:      userUsecase,
		accountUsecase:   accountUsecase,
		sessionUsecase:   sessionUsecase,
		oauthUsecase:     oauthUsecase,
		tokenService:     tokenService,
		logger:           logger,
	}
}

func (s *oauthService) GetProviders(ctx context.Context) []iamdomain.ProviderInfo {
	return []iamdomain.ProviderInfo{
		{Name: "google", DisplayName: "Google", Icon: "google"},
		{Name: "facebook", DisplayName: "Facebook", Icon: "facebook"},
	}
}

func (s *oauthService) InitiateLogin(ctx context.Context, provider iamdomain.ProviderType) (string, *http.Cookie, error) {
	p, ok := s.providerRegistry.Get(string(provider))
	if !ok {
		return "", nil, errors.BadRequestError("oauth.errors.unsupportedProvider")
	}

	state, cookie, err := s.stateManager.GenerateState(ctx, string(provider), false, "")
	if err != nil {
		return "", nil, errors.InternalError("oauth.errors.stateGenerationFailed", err)
	}

	session, err := p.BeginAuth(state)
	if err != nil {
		return "", nil, errors.InternalError("oauth.errors.authInitFailed", err)
	}

	s.stateManager.StoreSessionData(state, session.Marshal())

	authURL, err := session.GetAuthURL()
	if err != nil || authURL == "" {
		return "", nil, errors.InternalError("oauth.errors.authInitFailed", err)
	}

	return authURL, cookie, nil
}

func (s *oauthService) InitiateLink(ctx context.Context, provider iamdomain.ProviderType, accountID uuid.UUID) (string, *http.Cookie, error) {
	p, ok := s.providerRegistry.Get(string(provider))
	if !ok {
		return "", nil, errors.BadRequestError("oauth.errors.unsupportedProvider")
	}

	session, err := p.BeginAuth("")
	if err != nil {
		return "", nil, errors.InternalError("oauth.errors.authInitFailed", err)
	}

	state, cookie, err := s.stateManager.GenerateState(ctx, string(provider), true, accountID.String())
	if err != nil {
		return "", nil, errors.InternalError("oauth.errors.stateGenerationFailed", err)
	}

	s.stateManager.StoreSessionData(state, session.Marshal())

	authURL, _ := session.GetAuthURL()
	return authURL, cookie, nil
}

func (s *oauthService) HandleCallback(ctx context.Context, provider iamdomain.ProviderType, code string, state string) (*iamdomain.OAuthCallbackResult, *iamdomain.EmailRequiredResult, error) {
	p, ok := s.providerRegistry.Get(string(provider))
	if !ok {
		return nil, nil, errors.BadRequestError("oauth.errors.unsupportedProvider")
	}

	stateData, err := s.stateManager.ValidateState(ctx, state)
	if err != nil {
		return nil, nil, err
	}

	if stateData.IsLinking {
		return nil, nil, errors.BadRequestError("oauth.errors.invalidFlow")
	}

	sessionData := s.stateManager.GetSessionData(state)
	if sessionData == "" {
		return nil, nil, errors.InternalError("oauth.errors.sessionNotFound", nil)
	}

	session, err := p.UnmarshalSession(sessionData)
	if err != nil {
		return nil, nil, errors.InternalError("oauth.errors.sessionUnmarshalFailed", err)
	}

	s.logger.Debug("session data retrieved", zap.String("sessionType", fmt.Sprintf("%T", session)))
	s.logger.Debug("session data", zap.String("data", sessionData))

	gothProvider := p.GetGothProvider()
	s.logger.Debug("goth provider type", zap.String("type", fmt.Sprintf("%T", gothProvider)))

	_, err = session.Authorize(gothProvider, url.Values{"code": []string{code}})
	if err != nil {
		return nil, nil, errors.InternalError("oauth.errors.tokenExchangeFailed", err)
	}

	user, err := p.FetchUser(session)
	if err != nil {
		return nil, nil, errors.InternalError("oauth.errors.userInfoFailed", err)
	}

	s.stateManager.DeleteState(state)

	userInfo := s.extractUserInfo(provider, user)

	if userInfo.Email == "" {
		return nil, &iamdomain.EmailRequiredResult{
			PartialUserInfo: *userInfo,
			State:           state,
		}, nil
	}

	return s.processOAuthUser(ctx, provider, userInfo)
}

func (s *oauthService) HandleLinkCallback(ctx context.Context, provider iamdomain.ProviderType, code string, state string) (*iamdomain.LinkResult, *iamdomain.EmailRequiredResult, error) {
	p, ok := s.providerRegistry.Get(string(provider))
	if !ok {
		return nil, nil, errors.BadRequestError("oauth.errors.unsupportedProvider")
	}

	stateData, err := s.stateManager.ValidateState(ctx, state)
	if err != nil {
		return nil, nil, err
	}

	if !stateData.IsLinking {
		return nil, nil, errors.BadRequestError("oauth.errors.invalidFlow")
	}

	accountID, err := uuid.Parse(stateData.AccountID)
	if err != nil {
		return nil, nil, errors.BadRequestError("oauth.errors.invalidAccount")
	}

	sessionData := s.stateManager.GetSessionData(state)
	if sessionData == "" {
		return nil, nil, errors.InternalError("oauth.errors.sessionNotFound", nil)
	}

	session, err := p.UnmarshalSession(sessionData)
	if err != nil {
		return nil, nil, errors.InternalError("oauth.errors.sessionUnmarshalFailed", err)
	}

	gothProvider := p.GetGothProvider()
	_, err = session.Authorize(gothProvider, url.Values{"code": []string{code}})
	if err != nil {
		return nil, nil, errors.InternalError("oauth.errors.tokenExchangeFailed", err)
	}

	user, err := p.FetchUser(session)
	if err != nil {
		return nil, nil, errors.InternalError("oauth.errors.userInfoFailed", err)
	}

	s.stateManager.DeleteState(state)

	userInfo := s.extractUserInfo(provider, user)

	if userInfo.Email == "" {
		return nil, &iamdomain.EmailRequiredResult{
			PartialUserInfo: *userInfo,
			State:           state,
		}, nil
	}

	identity, err := s.oauthUsecase.LinkOAuthIdentity(ctx, accountID, usecase.LinkOAuthIdentityInput{
		Provider:        string(provider),
		ProviderSubject: userInfo.Subject,
		ProviderEmail:   &userInfo.Email,
	})
	if err != nil {
		return nil, nil, err
	}

	s.logger.Info("OAuth identity linked to account",
		core.String("identityID", identity.ID.String()),
		core.String("accountID", accountID.String()),
		core.String("provider", string(provider)),
	)

	return &iamdomain.LinkResult{OAuthIdentity: identity}, nil, nil
}

func (s *oauthService) CompleteWithEmail(ctx context.Context, state string, email string) (*iamdomain.OAuthCallbackResult, error) {
	return nil, errors.InternalError("oauth.errors.notImplemented", nil)
}

func (s *oauthService) GetLinkedIdentities(ctx context.Context, accountID uuid.UUID) ([]*entity.OAuthIdentity, error) {
	return s.oauthUsecase.ListOAuthIdentities(ctx, accountID)
}

func (s *oauthService) UnlinkProvider(ctx context.Context, accountID uuid.UUID, provider string) error {
	identities, err := s.oauthUsecase.ListOAuthIdentities(ctx, accountID)
	if err != nil {
		return err
	}

	for _, identity := range identities {
		if identity.Provider == provider {
			return s.oauthUsecase.UnlinkOAuthIdentity(ctx, accountID, identity.ID)
		}
	}

	return iamerror.ErrOAuthIdentityNotFound
}

func (s *oauthService) processOAuthUser(ctx context.Context, provider iamdomain.ProviderType, userInfo *iamdomain.OAuthUserInfo) (*iamdomain.OAuthCallbackResult, *iamdomain.EmailRequiredResult, error) {
	existingIdentities, _ := s.oauthUsecase.ListOAuthIdentities(ctx, uuid.Nil)
	for _, ident := range existingIdentities {
		if ident.Provider == string(provider) && ident.ProviderSubject == userInfo.Subject {
			result, err := s.loginExistingUser(ctx, ident)
			return result, nil, err
		}
	}

	existingAccount, err := s.accountUsecase.GetAccountByEmail(ctx, strings.ToLower(userInfo.Email))
	if err == nil && existingAccount != nil {
		_, err := s.oauthUsecase.LinkOAuthIdentity(ctx, existingAccount.ID, usecase.LinkOAuthIdentityInput{
			Provider:        string(provider),
			ProviderSubject: userInfo.Subject,
			ProviderEmail:   &userInfo.Email,
		})
		if err != nil {
			return nil, nil, err
		}

		user, err := s.userUsecase.GetUser(ctx, existingAccount.UserID)
		if err != nil {
			return nil, nil, err
		}

		result, err := s.generateTokens(ctx, user, existingAccount)
		if err != nil {
			return nil, nil, err
		}

		s.logger.Info("OAuth auto-linked to existing account",
			core.String("accountID", existingAccount.ID.String()),
			core.String("provider", string(provider)),
		)

		return &iamdomain.OAuthCallbackResult{
			IsNewUser:    false,
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresAt:    result.ExpiresAt,
			User:         result.User,
			Account:      result.Account,
		}, nil, nil
	}

	if err != nil && err != iamerror.ErrAccountNotFound {
		return nil, nil, err
	}

	result, err := s.createNewOAuthUser(ctx, provider, userInfo)
	return result, nil, err
}

func (s *oauthService) loginExistingUser(ctx context.Context, identity *entity.OAuthIdentity) (*iamdomain.OAuthCallbackResult, error) {
	account, err := s.accountUsecase.GetAccount(ctx, identity.AccountID)
	if err != nil {
		return nil, err
	}

	user, err := s.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return nil, err
	}

	result, err := s.generateTokens(ctx, user, account)
	if err != nil {
		return nil, err
	}

	return &iamdomain.OAuthCallbackResult{
		IsNewUser:    false,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		User:         result.User,
		Account:      account,
	}, nil
}

func (s *oauthService) createNewOAuthUser(ctx context.Context, provider iamdomain.ProviderType, userInfo *iamdomain.OAuthUserInfo) (*iamdomain.OAuthCallbackResult, error) {
	names := parseName(userInfo.Name)

	imageURL := userInfo.PictureURL
	user, err := s.userUsecase.CreateUser(ctx, usecase.CreateUserInput{
		FirstName: names.first,
		LastName:  names.last,
		ImageURL:  &imageURL,
	})
	if err != nil {
		return nil, err
	}

	account, err := s.accountUsecase.CreateAccount(ctx, usecase.CreateAccountInput{
		UserID:        user.ID,
		Email:         userInfo.Email,
		PasswordHash:  nil,
		EmailVerified: userInfo.EmailVerified,
		Status:        entity.AccountStatusActive,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.oauthUsecase.LinkOAuthIdentity(ctx, account.ID, usecase.LinkOAuthIdentityInput{
		Provider:        string(provider),
		ProviderSubject: userInfo.Subject,
		ProviderEmail:   &userInfo.Email,
	})
	if err != nil {
		return nil, err
	}

	result, err := s.generateTokens(ctx, user, account)
	if err != nil {
		return nil, err
	}

	s.logger.Info("New OAuth user created",
		core.String("userID", user.ID.String()),
		core.String("accountID", account.ID.String()),
		core.String("provider", string(provider)),
	)

	return &iamdomain.OAuthCallbackResult{
		IsNewUser:    true,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		User:         user,
		Account:      account,
	}, nil
}

func (s *oauthService) generateTokens(ctx context.Context, user *entity.User, account *entity.Account) (*service.AuthResult, error) {
	rawRefreshToken, refreshTokenHash, err := s.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionUsecase.CreateSession(ctx, account.ID, usecase.CreateSessionInput{
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(s.tokenService.GetRefreshTokenTTL()),
	})
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokenService.GenerateAccessToken(ctx, token.AccessTokenClaims{
		SessionID: session.ID,
		Email:     account.Email,
	})
	if err != nil {
		return nil, err
	}

	return &service.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    time.Now().Add(s.tokenService.GetAccessTokenTTL()),
		User:         user,
		Account:      account,
	}, nil
}

func (s *oauthService) extractUserInfo(provider iamdomain.ProviderType, user goth.User) *iamdomain.OAuthUserInfo {
	emailVerified := false

	switch provider {
	case iamdomain.ProviderGoogle:
		if verified, ok := user.RawData["email_verified"].(bool); ok {
			emailVerified = verified
		}
	case iamdomain.ProviderFacebook:
		emailVerified = true
	}

	return &iamdomain.OAuthUserInfo{
		Provider:      provider,
		Subject:       user.UserID,
		Email:         user.Email,
		EmailVerified: emailVerified,
		Name:          user.Name,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		PictureURL:    user.AvatarURL,
	}
}

type parsedName struct {
	first, last string
}

func parseName(fullName string) parsedName {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return parsedName{first: "Unknown"}
	}
	if len(parts) == 1 {
		return parsedName{first: parts[0]}
	}
	return parsedName{
		first: strings.Join(parts[:len(parts)-1], " "),
		last:  parts[len(parts)-1],
	}
}

var (
	_ = google.Session{}
	_ = facebook.Session{}
)
