package iam

import (
	"context"

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
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
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
			infrarepo.NewAccountEmailOTPRepository,
			fx.As(new(repository.AccountEmailOTPRepository)),
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
	fx.Provide(
		fx.Annotate(
			infrarepo.NewRoleRepository,
			fx.As(new(repository.RoleRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewPermissionRepository,
			fx.As(new(repository.PermissionRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewRolePermissionRepository,
			fx.As(new(repository.RolePermissionRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewRoleAssignmentRepository,
			fx.As(new(repository.RoleAssignmentRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewNotificationPreferenceRepository,
			fx.As(new(repository.NotificationPreferenceRepository)),
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
			appusecase.NewAccountEmailOTPUsecase,
			fx.As(new(usecase.AccountEmailOTPUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewSessionUsecase,
			fx.As(new(usecase.SessionUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewRoleUsecase,
			fx.As(new(usecase.RoleUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewPermissionUsecase,
			fx.As(new(usecase.PermissionUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewRoleAssignmentUsecase,
			fx.As(new(usecase.RoleAssignmentUsecase)),
		),
	),
	fx.Provide(service.NewRolePermissionSeeder),
	fx.Provide(service.NewSuperAdminSeeder),

	// Application Layer - Auth Service
	fx.Provide(
		fx.Annotate(
			service.NewAuthService,
			fx.As(new(service.AuthService)),
		),
	),
	fx.Provide(service.NewAdminService),

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
	fx.Provide(handler.NewAdminHandler),
	fx.Provide(handler.NewPermissionHandler),
	fx.Provide(handler.NewRoleHandler),
	fx.Provide(handler.NewUserHandler),
	fx.Provide(handler.NewImageHandler),
	fx.Provide(handler.NewOAuthHandler),

	// Invocations

	// Register routes
	fx.Invoke(func(api huma.API, authHandler *handler.AuthHandler, adminHandler *handler.AdminHandler, permissionHandler *handler.PermissionHandler, roleHandler *handler.RoleHandler, userHandler *handler.UserHandler, imageHandler *handler.ImageHandler, oauthHandler *handler.OAuthHandler, tokenService token.TokenService, authService service.AuthService, roleAssignmentUsecase usecase.RoleAssignmentUsecase) {
		authMiddleware := middleware.AuthMiddleware(api, tokenService, authService)
		AccountStatusMiddleware := middleware.AccountStatusMiddleware(api, authService)
		readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.ReadAccess, nil)
		writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.WriteAccess, nil)
		updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.UpdateAccess, nil)
		deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.DeleteAccess, nil)
		routes.RegisterRoutes(api, routes.RouteDependencies{
			AuthHandler:                authHandler,
			AdminHandler:               adminHandler,
			PermissionHandler:          permissionHandler,
			RoleHandler:                roleHandler,
			UserHandler:                userHandler,
			ImageHandler:               imageHandler,
			OAuthHandler:               oauthHandler,
			AuthMiddleware:             authMiddleware,
			AccountStatusMiddleware:    AccountStatusMiddleware,
			ReadPermissionMiddleware:   readPermissionMiddleware,
			WritePermissionMiddleware:  writePermissionMiddleware,
			UpdatePermissionMiddleware: updatePermissionMiddleware,
			DeletePermissionMiddleware: deletePermissionMiddleware,
		})
	}),

	// Register Google OAuth provider in registry
	fx.Invoke(func(registry *iamoauth.ProviderRegistry, googleProvider *iamoauth.GoogleProvider) {
		if googleProvider != nil {
			registry.Register(googleProvider)
		}
	}),

	// Seed IAM permissions and roles
	fx.Invoke(func(lc fx.Lifecycle, seeder *service.RolePermissionSeeder) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return seeder.Seed(ctx)
			},
		})
	}),

	// Seed super admin account
	fx.Invoke(func(lc fx.Lifecycle, seeder *service.SuperAdminSeeder) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return seeder.Seed(ctx)
			},
		})
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
