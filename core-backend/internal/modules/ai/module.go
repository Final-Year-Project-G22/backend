package ai

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	aiinfraclient "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/client"
	aiinfrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/repository"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	iammiddleware "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module("ai",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	fx.Provide(
		fx.Annotate(
			aiinfraclient.NewInferenceGRPCClient,
			fx.As(new(port.AIInferencePort)),
		),
	),
	fx.Provide(
		fx.Annotate(
			aiinfrarepo.NewIngestionDocumentRepository,
			fx.As(new(airepository.IngestionDocumentRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			aiinfrarepo.NewIngestionOutboxRepository,
			fx.As(new(airepository.IngestionOutboxRepository)),
		),
	),
	fx.Provide(service.NewIngestionService),
	fx.Provide(handler.NewIngestionHandler),
	fx.Invoke(func(api huma.API, ingestionHandler *handler.IngestionHandler, tokenService token.TokenService, authService iamservice.AuthService) {
		authMiddleware := iammiddleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := iammiddleware.AccountStatusMiddleware(api, authService)
		routes.RegisterRoutes(api, routes.RouteDependencies{
			IngestionHandler:        ingestionHandler,
			AuthMiddleware:          authMiddleware,
			AccountStatusMiddleware: accountStatusMiddleware,
		})
	}),
)
