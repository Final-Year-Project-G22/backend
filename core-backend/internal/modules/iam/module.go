package iam

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	iamoauth "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/oauth"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/repository"
	infratoken "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/token"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module("iam",
	// Entity Provider (for schema manager / migrations)
	fx.Provide(NewEntityProvider),

	// Register schema provider
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	// Transactor (core.Database implements sharedrepo.Transactor)
	fx.Provide(
		fx.Annotate(
			func(db *core.Database) sharedrepo.Transactor { return db },
			fx.As(new(sharedrepo.Transactor)),
		),
	),

	// Infrastructure Layer - Repositories
	fx.Provide(
		fx.Annotate(
			infrarepo.NewUserRepository,
			fx.As(new(repository.UserRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewAccountRepository,
			fx.As(new(repository.AccountRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewSessionRepository,
			fx.As(new(repository.SessionRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewOAuthIdentityRepository,
			fx.As(new(repository.OAuthIdentityRepository)),
		),
	),

	// OAuth Infrastructure
	fx.Provide(fx.Annotate(
		func(cfg *core.Config) (string, error) {
			return cfg.OAuth.EncryptionKey, nil
		},
		fx.ResultTags(`name:"oauthEncryptionKey"`),
	)),
	fx.Provide(fx.Annotate(
		iamoauth.NewTokenEncryptor,
		fx.ParamTags(`name:"oauthEncryptionKey"`),
	)),
	fx.Provide(fx.Annotate(
		func(cfg *core.Config) (string, bool) {
			return cfg.OAuth.CookieDomain, cfg.IsProduction()
		},
		fx.ResultTags(`name:"oauthCookieDomain"`, `name:"isProduction"`),
	)),
	fx.Provide(fx.Annotate(
		iamoauth.NewStateManager,
		fx.ParamTags(`name:"oauthCookieDomain"`, `name:"isProduction"`),
	)),
	fx.Provide(iamoauth.NewProviderRegistry),
	fx.Provide(fx.Annotate(
		func(cfg *core.Config) *core.OAuthProviderConfig {
			for i := range cfg.OAuth.Providers {
				if cfg.OAuth.Providers[i].Name == "google" {
					return &cfg.OAuth.Providers[i]
				}
			}
			return nil
		},
		fx.ResultTags(`name:"googleOAuthConfig"`),
	)),
	fx.Provide(fx.Annotate(
		iamoauth.NewGoogleProvider,
		fx.ParamTags(`name:"googleOAuthConfig"`),
	)),
	fx.Provide(fx.Annotate(
		func(cfg *core.Config) *core.OAuthProviderConfig {
			for i := range cfg.OAuth.Providers {
				if cfg.OAuth.Providers[i].Name == "facebook" {
					return &cfg.OAuth.Providers[i]
				}
			}
			return nil
		},
		fx.ResultTags(`name:"facebookOAuthConfig"`),
	)),
	fx.Provide(fx.Annotate(
		iamoauth.NewFacebookProvider,
		fx.ParamTags(`name:"facebookOAuthConfig"`),
	)),

	// Infrastructure Layer - Token Service
	fx.Provide(
		fx.Annotate(
			infratoken.NewJWTService,
			fx.As(new(token.TokenService)),
		),
	),

	// Application Layer - Usecases
	fx.Provide(
		fx.Annotate(
			appusecase.NewUserUsecase,
			fx.As(new(usecase.UserUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewAccountUsecase,
			fx.As(new(usecase.AccountUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewSessionUsecase,
			fx.As(new(usecase.SessionUsecase)),
		),
	),

	// Application Layer - Auth Service
	fx.Provide(
		fx.Annotate(
			service.NewAuthService,
			fx.As(new(service.AuthService)),
		),
	),

	// Application Layer - Avatar Service
	fx.Provide(service.NewAvatarValidator),
	fx.Provide(service.NewAvatarService),

	// Application Layer - OAuth Usecase
	fx.Provide(
		fx.Annotate(
			appusecase.NewOAuthIdentityUsecase,
			fx.As(new(usecase.OAuthIdentityUsecase)),
		),
	),

	// OAuth Service
	fx.Provide(iamoauth.NewOAuthService),

	// Delivery Layer - Handler
	fx.Provide(handler.NewAuthHandler),
	fx.Provide(handler.NewUserHandler),
	fx.Provide(handler.NewImageHandler),
	fx.Provide(handler.NewOAuthHandler),

	// Invocations

	// Register routes
	fx.Invoke(func(api huma.API, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, imageHandler *handler.ImageHandler, oauthHandler *handler.OAuthHandler, tokenService token.TokenService, authService service.AuthService) {
		authMiddleware := middleware.AuthMiddleware(api, tokenService, authService)
		routes.RegisterRoutes(api, routes.RouteDependencies{
			AuthHandler:    authHandler,
			UserHandler:    userHandler,
			ImageHandler:   imageHandler,
			OAuthHandler:   oauthHandler,
			AuthMiddleware: authMiddleware,
		})
	}),

	// Register Google OAuth provider in registry
	fx.Invoke(func(registry *iamoauth.ProviderRegistry, googleProvider *iamoauth.GoogleProvider) {
		if googleProvider != nil {
			registry.Register(googleProvider)
		}
	}),
)

// TransactorProvider provides the core.Database as a sharedrepo.Transactor.
// This is needed because service.NewAuthService expects sharedrepo.Transactor.
// The core.Database already implements this interface, so we just need to
// tell fx how to provide it.
func init() {
	// This is handled automatically by fx since core.Database implements
	// sharedrepo.Transactor. If needed, we can add an explicit provider:
	// fx.Provide(func(db *core.Database) sharedrepo.Transactor { return db })
}

// Ensure core.Database satisfies sharedrepo.Transactor at compile time.
var _ sharedrepo.Transactor = (*core.Database)(nil)
