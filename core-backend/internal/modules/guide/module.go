package guide

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/infrastructure/repository"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	iammiddleware "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"guide",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	fx.Provide(
		fx.Annotate(
			infrarepo.NewCategoryRepository,
			fx.As(new(repository.CategoryRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewGuideRepository,
			fx.As(new(repository.GuideRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewStepRepository,
			fx.As(new(repository.StepRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewProgressRepository,
			fx.As(new(repository.ProgressRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			appusecase.NewGuideViewUsecase,
			fx.As(new(usecase.GuideViewUseCase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewGuideAdminUsecase,
			fx.As(new(usecase.GuideManagementUseCase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewJourneyManagementUsecase,
			fx.As(new(usecase.JourneyManagementUseCase)),
		),
	),

	fx.Provide(handler.NewGuideViewHandler),

	fx.Invoke(func(api huma.API, guideViewHandler *handler.GuideViewHandler, tokenService token.TokenService, authService iamservice.AuthService) {
		authMiddleware := iammiddleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := iammiddleware.AccountStatusMiddleware(api, authService)
		routes.RegisterRoutes(api, routes.RouteDependencies{
			GuideViewHandler:        guideViewHandler,
			AuthMiddleware:          authMiddleware,
			AccountStatusMiddleware: accountStatusMiddleware,
		})
	}),
)
