package main

import (
	"encoding/json"
	"os"

	aihandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	airoutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/routes"
	communityhandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/handler"
	communityroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/routes"
	guidehandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/handler"
	guideroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/routes"
	iamhandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	iamroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/routes"
	libraryhandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/delivery/handler"
	libraryroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/delivery/routes"
	notificationhandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	notificationroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/routes"
	paymenthandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/delivery/handler"
	paymentroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/delivery/routes"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.New()

	config := huma.DefaultConfig("Adisu Backend API", "1.0.0")
	config.Info.Description = "Backend API for the Adisu platform"

	api := humagin.New(engine, config)

	noOpMiddleware := func(ctx huma.Context, next func(huma.Context)) {
		next(ctx)
	}

	iamroutes.RegisterRoutes(api, iamroutes.RouteDependencies{
		AuthHandler:             &iamhandler.AuthHandler{},
		AdminHandler:            &iamhandler.AdminHandler{},
		PermissionHandler:       &iamhandler.PermissionHandler{},
		RoleHandler:             &iamhandler.RoleHandler{},
		UserHandler:             &iamhandler.UserHandler{},
		ImageHandler:            &iamhandler.ImageHandler{},
		OAuthHandler:            &iamhandler.OAuthHandler{},
		AuthMiddleware:          noOpMiddleware,
		AccountStatusMiddleware: noOpMiddleware,
		RoleAssignmentUsecase:   nil,
	})

	communityroutes.RegisterRoutes(api, communityroutes.RouteDependencies{
		CommunityHandler:           &communityhandler.CommunityHandler{},
		CommunityAdminHandler:      &communityhandler.CommunityAdminHandler{},
		AuthMiddleware:             noOpMiddleware,
		AccountStatusMiddleware:    noOpMiddleware,
		ReadPermissionMiddleware:   noOpMiddleware,
		WritePermissionMiddleware:  noOpMiddleware,
		UpdatePermissionMiddleware: noOpMiddleware,
		DeletePermissionMiddleware: noOpMiddleware,
	})

	airoutes.RegisterRoutes(api, airoutes.RouteDependencies{
		IngestionHandler:        &aihandler.IngestionHandler{},
		StatusHandler:           &aihandler.StatusHandler{},
		AskHandler:              &aihandler.AskHandler{},
		DLQHandler:              &aihandler.DLQHandler{},
		SSEHandler:              &aihandler.SSEHandler{},
		ToggleHandler:           &aihandler.ToggleHandler{},
		AskEnabled:              true,
		AuthMiddleware:          noOpMiddleware,
		AccountStatusMiddleware: noOpMiddleware,
	})

	notificationroutes.RegisterRoutes(api, engine, notificationroutes.RouteDependencies{
		AdminHandler:            &notificationhandler.NotificationAdminHandler{},
		NotificationHandler:     &notificationhandler.NotificationHandler{},
		WebhookHandler:          &notificationhandler.WebhookHandler{},
		ComplianceHandler:       &notificationhandler.ComplianceHandler{},
		AuthMiddleware:          noOpMiddleware,
		AccountStatusMiddleware: noOpMiddleware,
	})

	libraryroutes.RegisterRoutes(api, libraryroutes.RouteDependencies{
		ViewHandler:                &libraryhandler.LibraryHandler{},
		AdminHandler:               &libraryhandler.LibraryAdminHandler{},
		AuthMiddleware:             noOpMiddleware,
		AccountStatusMiddleware:    noOpMiddleware,
		ReadPermissionMiddleware:   noOpMiddleware,
		WritePermissionMiddleware:  noOpMiddleware,
		UpdatePermissionMiddleware: noOpMiddleware,
		DeletePermissionMiddleware: noOpMiddleware,
	})

	guideroutes.RegisterRoutes(api, guideroutes.RouteDependencies{
		GuideViewHandler:        &guidehandler.GuideViewHandler{},
		GuideAdminHandler:       &guidehandler.GuideAdminHandler{},
		AuthMiddleware:          noOpMiddleware,
		AccountStatusMiddleware: noOpMiddleware,
	})

	paymentroutes.RegisterPaymentRoutes(api, engine, paymentroutes.RouteDependencies{
		PaymentHandler:          &paymenthandler.PaymentHandler{},
		WebhookHandler:          &paymenthandler.WebhookHandler{},
		AuthMiddleware:          noOpMiddleware,
		AccountStatusMiddleware: noOpMiddleware,
	})

	file, err := os.Create("docs/openapi.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(api.OpenAPI()); err != nil {
		panic(err)
	}
}
