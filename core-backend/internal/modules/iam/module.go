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
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/repository"
	infratoken "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/token"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module("iam",
	// Entity Provider (for schema manager / migrations)
	fx.Provide(
		fx.Annotate(
			NewEntityProvider,
			fx.As(new(core.EntityProvider)),
		),
	),

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

	// Delivery Layer - Handler
	fx.Provide(handler.NewAuthHandler),

	// Invocations

	// Register schema provider
	fx.Invoke(func(sm *core.SchemaManager, provider core.EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	// Register auth routes
	fx.Invoke(func(api huma.API, authHandler *handler.AuthHandler, tokenService token.TokenService, authService service.AuthService) {
		authMiddleware := middleware.AuthMiddleware(api, tokenService, authService)
		routes.RegisterRoutes(api, routes.RouteDependencies{
			AuthHandler:    authHandler,
			AuthMiddleware: authMiddleware,
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
