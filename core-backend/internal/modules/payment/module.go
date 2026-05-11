package payment

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	iamtoken "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/delivery/routes"
	paymentrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/repository"
	paymentuc "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/infrastructure/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/chapa"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module is the payment module DI container.
var Module = fx.Module("payment",
	// Entity Provider (for schema manager / migrations)
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	// Infrastructure Layer - Chapa Client
	fx.Provide(func(cfg *core.Config) chapa.Client {
		if !cfg.Chapa.Enabled {
			return nil
		}
		return chapa.NewClient(chapa.Config{
			SecretKey: cfg.Chapa.SecretKey,
			BaseURL:   cfg.Chapa.BaseURL,
		})
	}),

	// Infrastructure Layer - Repositories
	fx.Provide(
		fx.Annotate(
			infrarepo.NewPlanRepository,
			fx.As(new(paymentrepo.PlanRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewPaymentRepository,
			fx.As(new(paymentrepo.PaymentRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewSubscriptionRepository,
			fx.As(new(paymentrepo.SubscriptionRepository)),
		),
	),

	// Application Layer - Usecases
	fx.Provide(
		fx.Annotate(
			usecase.NewPaymentUseCase,
			fx.As(new(paymentuc.PaymentUseCase)),
		),
	),

	// Delivery Layer - Handlers
	fx.Provide(handler.NewPaymentHandler),
	fx.Provide(handler.NewWebhookHandler),

	// Invocations
	fx.Invoke(func(
		api huma.API,
		engine *gin.Engine,
		paymentHandler *handler.PaymentHandler,
		webhookHandler *handler.WebhookHandler,
		tokenService iamtoken.TokenService,
		authService iamservice.AuthService,
	) {
		authMiddleware := middleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := middleware.AccountStatusMiddleware(api, authService)
		routes.RegisterPaymentRoutes(api, engine, routes.RouteDependencies{
			PaymentHandler:          paymentHandler,
			WebhookHandler:          webhookHandler,
			AuthMiddleware:          authMiddleware,
			AccountStatusMiddleware: accountStatusMiddleware,
		})
	}),
)
