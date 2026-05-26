package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	ai_tool_port "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/domain/port"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	appservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	emailProvider "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/email"
	pushProvider "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/push"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/infrastructure/repository"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
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
	fx.Provide(fx.Annotate(infrarepo.NewCampaignTemplateRepository, fx.As(new(repository.CampaignTemplateRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserNotificationPreferenceRepository, fx.As(new(repository.UserNotificationPreferenceRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewMutedAccountRepository, fx.As(new(repository.MutedAccountRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserDeviceRepository, fx.As(new(repository.UserDeviceRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationQueueRepository, fx.As(new(repository.NotificationQueueRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationHistoryRepository, fx.As(new(repository.NotificationHistoryRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserNotificationInboxRepository, fx.As(new(repository.UserNotificationInboxRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewEmailDeliveryLogRepository, fx.As(new(repository.EmailDeliveryLogRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationCampaignRepository, fx.As(new(repository.NotificationCampaignRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewNotificationOutboxRepository, fx.As(new(repository.NotificationOutboxRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewUserScheduledNotificationRepository, fx.As(new(repository.UserScheduledNotificationRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewScheduledAlertTemplateRepository, fx.As(new(repository.ScheduledAlertTemplateRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewComplianceEntryRepository, fx.As(new(repository.ComplianceEntryRepository)))),
	fx.Provide(fx.Annotate(infrarepo.NewComplianceTypeRepository, fx.As(new(repository.ComplianceTypeRepository)))),
	fx.Provide(infrarepo.NewThreadMuteResolver),

	// --- Services ---
	fx.Provide(appservice.NewTemplateRenderer),
	fx.Provide(appservice.NewCampaignSSEBroadcaster),
	fx.Provide(appservice.NewInboxSSEBroadcaster),
	fx.Provide(appservice.NewDeliveryWorker),
	fx.Provide(appservice.NewCampaignProcessor),
	fx.Provide(appservice.NewCampaignScheduler),
	fx.Provide(appservice.NewNotificationOutboxDispatcher),
	fx.Provide(appservice.NewUserNotificationScheduler),
	fx.Provide(appservice.NewBusinessAlertScheduler),
	fx.Provide(appservice.NewSyncComplianceService),

	// --- Email Provider ---
	fx.Provide(fx.Annotate(func(cfg *core.Config, logger core.Logger) repository.EmailProvider {
		if cfg.Email.Enabled {
			return emailProvider.NewSMTPProvider(cfg.Email, logger)
		}
		if cfg.Resend.Enabled {
			return emailProvider.NewResendProvider(cfg.Resend, logger)
		}
		return nil
	}, fx.As(new(repository.EmailProvider)))),

	// --- Push Provider ---
	fx.Provide(fx.Annotate(func(cfg *core.Config, logger core.Logger) repository.PushProvider {
		if cfg.FCM.CredentialsFile != "" {
			provider, err := pushProvider.NewFCMProvider(cfg.FCM, logger)
			if err != nil {
				logger.Warn("Failed to initialize FCM provider, falling back to noop",
					core.String("credentialsFile", cfg.FCM.CredentialsFile),
					core.Error(err),
				)
				return pushProvider.NewNoopProvider(logger)
			}
			return provider
		}
		return pushProvider.NewNoopProvider(logger)
	}, fx.As(new(repository.PushProvider)))),

	// --- IAM global preference reader (default: enabled) ---
	fx.Provide(fx.Annotate(func() appusecase.IAMGlobalPreferenceReader {
		return &defaultIAMReader{}
	}, fx.As(new(appusecase.IAMGlobalPreferenceReader)))),

	// --- Account reader (default: empty — replace with IAM provider) ---
	fx.Provide(fx.Annotate(func() repository.AccountReader {
		return &defaultAccountReader{}
	}, fx.As(new(repository.AccountReader)))),

	// --- Subscription reader (default: not pro — replace with payment provider) ---
	fx.Provide(fx.Annotate(func() repository.SubscriptionReader {
		return &defaultSubscriptionReader{}
	}, fx.As(new(repository.SubscriptionReader)))),

	// --- Use Cases ---
	fx.Provide(fx.Annotate(appusecase.NewNotificationTemplateUsecase, fx.As(new(usecase.NotificationTemplateUsecase)))),
	fx.Provide(fx.Annotate(func(
		tmplRepo repository.NotificationTemplateRepository,
		prefRepo repository.UserNotificationPreferenceRepository,
		mutedRepo repository.MutedAccountRepository,
		queueRepo repository.NotificationQueueRepository,
		accountRepo iamrepo.AccountRepository,
		iamReader appusecase.IAMGlobalPreferenceReader,
		renderer *appservice.TemplateRenderer,
		transactor sharedrepo.Transactor,
		muteResolver repository.MuteResolver,
	) usecase.NotificationIngestUsecase {
		resolvers := []repository.MuteResolver{muteResolver}
		return appusecase.NewNotificationIngestUsecase(
			tmplRepo, prefRepo, mutedRepo, queueRepo, accountRepo,
			iamReader, renderer, transactor,
			resolvers,
		)
	}, fx.As(new(usecase.NotificationIngestUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationDeliveryUsecase, fx.As(new(usecase.NotificationDeliveryUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationInboxUsecase, fx.As(new(usecase.NotificationInboxUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationHistoryUsecase, fx.As(new(usecase.NotificationHistoryUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationPreferenceUsecase, fx.As(new(usecase.NotificationPreferenceUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationMuteUsecase, fx.As(new(usecase.NotificationMuteUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationDeviceUsecase, fx.As(new(usecase.NotificationDeviceUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewEmailDeliveryUsecase, fx.As(new(usecase.EmailDeliveryUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewCampaignTemplateUsecase, fx.As(new(usecase.CampaignTemplateUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewNotificationCampaignUsecase, fx.As(new(usecase.NotificationCampaignUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewUserScheduledNotificationUsecase, fx.As(new(usecase.UserScheduledNotificationUsecase)))),
	fx.Provide(fx.Annotate(appusecase.NewComplianceEntryUsecase, fx.As(new(usecase.ComplianceEntryUsecase)))),

	// AI tool handlers
	fx.Provide(
		fx.Annotate(
			NewCheckComplianceStatusTool,
			fx.As(new(ai_tool_port.ToolHandler)),
			fx.ResultTags(`group:"ai_tool_handlers"`),
		),
	),

	// --- Handlers ---
	fx.Provide(handler.NewNotificationAdminHandler),
	fx.Provide(handler.NewCampaignTemplateHandler),
	fx.Provide(handler.NewNotificationHandler),
	fx.Provide(handler.NewWebhookHandler),
	fx.Provide(handler.NewSSEHandler),
	fx.Provide(handler.NewInboxSSEHandler),
	fx.Provide(handler.NewScheduledAlertHandler),
	fx.Provide(handler.NewComplianceHandler),

	fx.Provide(
		fx.Annotate(
			NotificationSeedPermissions,
			fx.ResultTags(`group:"permission_seeds"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			NotificationSeedRoles,
			fx.ResultTags(`group:"role_seeds"`),
		),
	),

	// --- Routes ---
	fx.Invoke(func(
		api huma.API,
		engine *gin.Engine,
		adminHandler *handler.NotificationAdminHandler,
		campaignTemplateHandler *handler.CampaignTemplateHandler,
		notificationHandler *handler.NotificationHandler,
		webhookHandler *handler.WebhookHandler,
		sseHandler *handler.SSEHandler,
		inboxSSEHandler *handler.InboxSSEHandler,
		scheduledAlertHandler *handler.ScheduledAlertHandler,
		complianceHandler *handler.ComplianceHandler,
		tokenService token.TokenService,
		authService service.AuthService,
		roleAssignmentUsecase iamusecase.RoleAssignmentUsecase,
	) {
		authMiddleware := middleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := middleware.AccountStatusMiddleware(api, authService)
		readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, NotificationRead, []string{"super_admin"})
		writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, NotificationWrite, []string{"super_admin"})
		updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, NotificationUpdate, []string{"super_admin"})
		deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, NotificationDelete, []string{"super_admin"})
		routes.RegisterRoutes(api, engine, routes.RouteDependencies{
			AdminHandler:               adminHandler,
			CampaignTemplateHandler:    campaignTemplateHandler,
			NotificationHandler:        notificationHandler,
			WebhookHandler:             webhookHandler,
			SSEHandler:                 sseHandler,
			InboxSSEHandler:            inboxSSEHandler,
			ScheduledAlertHandler:      scheduledAlertHandler,
			ComplianceHandler:          complianceHandler,
			AuthMiddleware:             authMiddleware,
			AccountStatusMiddleware:    accountStatusMiddleware,
			ReadPermissionMiddleware:   readPermissionMiddleware,
			WritePermissionMiddleware:  writePermissionMiddleware,
			UpdatePermissionMiddleware: updatePermissionMiddleware,
			DeletePermissionMiddleware: deletePermissionMiddleware,
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

	// --- Campaign SSE Broadcaster ---
	fx.Invoke(func(lc fx.Lifecycle, broadcaster *appservice.CampaignSSEBroadcaster) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				broadcaster.Start(ctx)
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),

	// --- Inbox SSE Broadcaster ---
	fx.Invoke(func(lc fx.Lifecycle, broadcaster *appservice.InboxSSEBroadcaster) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				broadcaster.Start(ctx)
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),

	// --- User Notification Scheduler ---
	fx.Invoke(func(lc fx.Lifecycle, scheduler *appservice.UserNotificationScheduler) {
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

	// --- Business Alert Scheduler ---
	fx.Invoke(func(lc fx.Lifecycle, scheduler *appservice.BusinessAlertScheduler) {
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

func (r *defaultAccountReader) GetAccountInfo(_ context.Context, _ uuid.UUID) (*repository.AccountInfo, error) {
	return &repository.AccountInfo{Email: "", Locale: "en", Name: ""}, nil
}

type defaultSubscriptionReader struct{}

func (r *defaultSubscriptionReader) HasActiveProSubscription(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}
