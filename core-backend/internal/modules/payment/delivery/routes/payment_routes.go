package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
)

const paymentBase = "/api/v1/payments"
const meBase = "/api/v1/me"
const webhookBase = "/api/v1/webhooks"

// RouteDependencies holds all dependencies needed to register payment routes.
type RouteDependencies struct {
	PaymentHandler          *handler.PaymentHandler
	WebhookHandler          *handler.WebhookHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

// RegisterPaymentRoutes registers all payment module routes.
func RegisterPaymentRoutes(api huma.API, engine *gin.Engine, deps RouteDependencies) {
	// --- Public: List Plans ---
	huma.Register(api, huma.Operation{
		OperationID: "listPlans",
		Method:      "GET",
		Path:        paymentBase + "/plans",
		Summary:     "List subscription plans",
		Description: "Returns all available subscription plans.",
		Tags:        []string{"Payments"},
	}, deps.PaymentHandler.HandleListPlans)

	// --- Authenticated: Initiate Payment ---
	huma.Register(api, huma.Operation{
		OperationID: "initiatePayment",
		Method:      "POST",
		Path:        paymentBase + "/initiate",
		Summary:     "Initiate payment",
		Description: "Starts a new payment transaction and returns a Chapa checkout URL.",
		Tags:        []string{"Payments"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.PaymentHandler.HandleInitiatePayment)

	// --- Authenticated: Verify Payment ---
	huma.Register(api, huma.Operation{
		OperationID: "verifyPayment",
		Method:      "POST",
		Path:        paymentBase + "/verify",
		Summary:     "Verify payment",
		Description: "Verifies the status of a payment by transaction reference.",
		Tags:        []string{"Payments"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.PaymentHandler.HandleVerifyPayment)

	// --- Authenticated: Get My Subscription ---
	huma.Register(api, huma.Operation{
		OperationID: "getMySubscription",
		Method:      "GET",
		Path:        meBase + "/subscription",
		Summary:     "Get my subscription",
		Description: "Returns the current account's active subscription.",
		Tags:        []string{"Payments"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.PaymentHandler.HandleGetMySubscription)

	// --- Webhook: Chapa (raw gin, no auth) ---
	engine.POST(webhookBase+"/chapa", deps.WebhookHandler.HandleChapaWebhook)
}
