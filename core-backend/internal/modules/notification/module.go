package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	appservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	emailProvider "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/email"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/repository"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"notification",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	// --- Repositories ---
	fx.Provide(fx.Annotate(infrarepo.NewNotificationTemplateRepository, fx.As(new(repository.NotificationTemplateRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserNotificationPreferenceRepository, fx.As(new(repository.UserNotificationPreferenceRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewMutedAccountRepository, fx.As(new(repository.MutedAccountRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserDeviceRepository, fx.As(new(repository.UserDeviceRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationQueueRepository, fx.As(new(repository.NotificationQueueRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationHistoryRepository, fx.As(new(repository.NotificationHistoryRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserNotificationInboxRepository, fx.As(new(repository.UserNotificationInboxRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewEmailDeliveryLogRepository, fx.As(new(repository.EmailDeliveryLogRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationCampaignRepository, fx.As(new(repository.NotificationCampaignRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationOutboxRepository, fx.As(new(repository.NotificationOutboxRepository)))),

	// --- Services ---
	fx.Provide(appservice.NewTemplateRenderer),
	fx.Provide(appservice.NewDeliveryWorker),
	fx.Provide(appservice.NewCampaignProcessor),
	fx.Provide(appservice.NewCampaignScheduler),
	fx.Provide(appservice.NewNotificationOutboxDispatcher),

	// --- Email Provider ---
	fx.Provide(fx.Annotate(func(cfg *core.Config, logger core.Logger) repository.EmailProvider {
		return emailProvider.NewResendProvider(cfg.Resend, logger)
	}, fx.As(new(repository.EmailProvider)))),

	// --- IAM global preference reader (default: enabled) ---
	fx.Provide(fx.Annotate(func() appusecase.IAMGlobalPreferenceReader {
		return &defaultIAMReader{}
	}, fx.As(new(appusecase.IAMGlobalPreferenceReader)))),

	// --- Account reader (default: empty — replace with IAM provider) ---
	fx.Provide(fx.Annotate(func() repository.AccountReader {
		return &defaultAccountReader{}
	}, fx.As(new(repository.AccountReader)))),

	// --- Use Cases ---
	fx.Provide(fx.Annotate(appusecase.NewNotificationTemplateUsecase, fx.As(new(usecase.NotificationTemplateUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationIngestUsecase, fx.As(new(usecase.NotificationIngestUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationDeliveryUsecase, fx.As(new(usecase.NotificationDeliveryUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationInboxUsecase, fx.As(new(usecase.NotificationInboxUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationHistoryUsecase, fx.As(new(usecase.NotificationHistoryUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationPreferenceUsecase, fx.As(new(usecase.NotificationPreferenceUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationMuteUsecase, fx.As(new(usecase.NotificationMuteUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationDeviceUsecase, fx.As(new(usecase.NotificationDeviceUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewEmailDeliveryUsecase, fx.As(new(usecase.EmailDeliveryUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationCampaignUsecase, fx.As(new(usecase.NotificationCampaignUsecase)))),

	// --- Handlers ---
	fx.Provide(handler.NewNotificationAdminHandler),
	fx.Provide(handler.NewNotificationHandler),
	fx.Provide(handler.NewWebhookHandler),

	// --- Routes ---
	fx.Invoke(func(
		api huma.API,
		engine *gin.Engine,
		adminHandler *handler.NotificationAdminHandler,
		notificationHandler *handler.NotificationHandler,
		webhookHandler *handler.WebhookHandler,
		tokenService token.TokenService,
		authService service.AuthService,
	) {
		authMiddleware := middleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := middleware.AccountStatusMiddleware(api, authService)
		routes.RegisterRoutes(api, engine, routes.RouteDependencies{
			AdminHandler:            adminHandler,
			NotificationHandler:     notificationHandler,
			WebhookHandler:          webhookHandler,
			AuthMiddleware:          authMiddleware,
			AccountStatusMiddleware: accountStatusMiddleware,
		})
	}),

	// --- Delivery Worker ---
	fx.Invoke(func(lc fx.Lifecycle, worker *appservice.DeliveryWorker) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				worker.Start(ctx)
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),

	// --- Campaign Scheduler ---
	fx.Invoke(func(lc fx.Lifecycle, scheduler *appservice.CampaignScheduler) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				scheduler.Start(ctx)
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),

	// --- Notification Outbox Dispatcher ---
	fx.Invoke(func(lc fx.Lifecycle, dispatcher *appservice.NotificationOutboxDispatcher) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go dispatcher.Start(ctx)
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),

	// --- Event Subscriptions ---
	fx.Invoke(registerEventSubscriptions),
)

type defaultIAMReader struct{}

func (r *defaultIAMReader) IsNotificationEnabled(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return true, nil
}

type defaultAccountReader struct{}

func (r *defaultAccountReader) FindAll(_ context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

func (r *defaultAccountReader) FindBySegment(_ context.Context, _ map[string]interface{}) ([]uuid.UUID, error) {
	return nil, nil
}
