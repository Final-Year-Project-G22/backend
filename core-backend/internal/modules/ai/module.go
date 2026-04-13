package ai

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	aiinfraclient "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/client"
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
	fx.Provide(service.NewIngestionService),
	fx.Provide(handler.NewIngestionHandler),
	fx.Invoke(func(api huma.API, ingestionHandler *handler.IngestionHandler) {
		routes.RegisterRoutes(api, routes.RouteDependencies{
			IngestionHandler: ingestionHandler,
		})
	}),
)
