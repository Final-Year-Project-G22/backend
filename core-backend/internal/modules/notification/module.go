package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	appservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	emailProvider "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/email"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/repository"
	"github.com/danielgtaylor/huma/v2"
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

	// --- Services ---
	fx.Provide(appservice.NewTemplateRenderer),
	fx.Provide(appservice.NewDeliveryWorker),

	// --- Email Provider ---
	fx.Provide(fx.Annotate(func(cfg *core.Config, logger core.Logger) emailProvider.EmailProvider {
		return emailProvider.NewResendProvider(cfg.Resend, logger)
	}, fx.As(new(emailProvider.EmailProvider)))),

	// --- IAM global preference reader (default: enabled) ---
	fx.Provide(fx.Annotate(func() appusecase.IAMGlobalPreferenceReader {
		return &defaultIAMReader{}
	}, fx.As(new(appusecase.IAMGlobalPreferenceReader)))),

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

	// --- Handlers ---
	fx.Provide(handler.NewNotificationAdminHandler),
	fx.Provide(handler.NewNotificationHandler),

	// --- Routes ---
	fx.Invoke(func(
		api huma.API,
		adminHandler *handler.NotificationAdminHandler,
		notificationHandler *handler.NotificationHandler,
	) {
		routes.RegisterRoutes(api, routes.RouteDependencies{
			AdminHandler:        adminHandler,
			NotificationHandler: notificationHandler,
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
)

type defaultIAMReader struct{}

func (r *defaultIAMReader) IsNotificationEnabled(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}
