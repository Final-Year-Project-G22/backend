package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	aisvc "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/service"
	aiinfra "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure"
	aiinfraclient "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/client"
	aiinframsg "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/messaging"
	aiinfrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/repository"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	iammiddleware "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	stg "github.com/Final-Year-Project-G22/backend/core/pkg/storage"
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
			fx.ResultTags(`name:"baseInferencePort"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			aiinfrarepo.NewConversationCache,
			fx.As(new(port.ConversationCachePort)),
		),
	),
	fx.Provide(
		fx.Annotate(
			func(
				cfg *core.Config,
				cache port.ConversationCachePort,
				base port.AIInferencePort,
			) *service.CachingInferencePort {
				return service.NewCachingInferencePort(base, cache, cfg.AI.ConversationCacheTTL)
			},
			fx.ParamTags(``, ``, `name:"baseInferencePort"`),
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
			aiinfrarepo.NewIngestionStatusProjectionRepository,
			fx.As(new(airepository.IngestionStatusProjectionRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			aisvc.NewEnvelopeSigner,
			fx.As(new(aisvc.EnvelopeSigner)),
		),
	),
	fx.Provide(func(cfg *core.Config, s stg.Storage, docRepo airepository.IngestionDocumentRepository, outboxRepo airepository.IngestionOutboxRepository, projectionRepo airepository.IngestionStatusProjectionRepository, transactor sharedrepo.Transactor) *service.IngestionService {
		return service.NewIngestionService(cfg.Ingestion.Enabled, s, docRepo, outboxRepo, projectionRepo, transactor)
	}),
	fx.Provide(service.NewOutboxDispatcher),
	fx.Provide(service.NewSSEGateway),
	fx.Provide(service.NewStatusProjectionService),
	fx.Provide(aiinframsg.NewStatusEventSubscriber),
	fx.Provide(handler.NewIngestionHandler),
	fx.Provide(handler.NewStatusHandler),
	fx.Provide(service.NewAskService),
	fx.Provide(handler.NewAskHandler),
	fx.Provide(handler.NewSSEHandler),
	fx.Provide(handler.NewToggleHandler),
	fx.Provide(
		fx.Annotate(
			AISeedPermissions,
			fx.ResultTags(`group:"permission_seeds"`),
		),
	),
	fx.Provide(func(dlqController port.DLQController) *handler.DLQHandler {
		return handler.NewDLQHandler(dlqController)
	}),
	fx.Provide(
		fx.Annotate(
			aiinfra.NewDLQController,
			fx.As(new(port.DLQController)),
		),
	),
	fx.Provide(
		fx.Annotate(
			aiinfra.NewIngestToggle,
			fx.As(new(port.IngestControl)),
		),
	),
	fx.Invoke(func(cfg *core.Config, api huma.API, ingestionHandler *handler.IngestionHandler, statusHandler *handler.StatusHandler, askHandler *handler.AskHandler, dlqHandler *handler.DLQHandler, sseHandler *handler.SSEHandler, toggleHandler *handler.ToggleHandler, tokenService token.TokenService, authService iamservice.AuthService, roleAssignmentUsecase iamusecase.RoleAssignmentUsecase, logger core.Logger) {
		authMiddleware := iammiddleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := iammiddleware.AccountStatusMiddleware(api, authService)
		adminPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, AIAdminStream, []string{"super_admin"})
		routes.RegisterRoutes(api, routes.RouteDependencies{
			IngestionHandler:          ingestionHandler,
			StatusHandler:             statusHandler,
			AskHandler:                askHandler,
			DLQHandler:                dlqHandler,
			SSEHandler:                sseHandler,
			ToggleHandler:             toggleHandler,
			Logger:                    logger,
			AskEnabled:                cfg.AI.AskEnabled,
			AuthMiddleware:            authMiddleware,
			AccountStatusMiddleware:   accountStatusMiddleware,
			AdminPermissionMiddleware: adminPermissionMiddleware,
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
					ticker := time.NewTicker(dispatcher.Interval())
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							_ = dispatcher.DispatchBatch(ctx, dispatcher.BatchSize())
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
	fx.Invoke(func(lc fx.Lifecycle, subscriber *aiinframsg.StatusEventSubscriber) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				if err := subscriber.Subscribe(); err != nil {
					return fmt.Errorf("failed to subscribe to status events: %w", err)
				}
				return nil
			},
			OnStop: func(context.Context) error {
				return nil
			},
		})
	}),
)
