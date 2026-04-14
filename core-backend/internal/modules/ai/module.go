package ai

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	aisvc "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/service"
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
	fx.Provide(
		fx.Annotate(
			aisvc.NewEnvelopeSigner,
			fx.As(new(aisvc.EnvelopeSigner)),
		),
	),
	fx.Provide(service.NewIngestionService),
	fx.Provide(service.NewOutboxDispatcher),
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
	fx.Invoke(func(lc fx.Lifecycle, cfg *core.Config, dispatcher *service.OutboxDispatcher) {
		if !cfg.Ingestion.Enabled {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go func() {
					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							_ = dispatcher.DispatchBatch(ctx, 50)
						}
					}
				}()
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),
)
