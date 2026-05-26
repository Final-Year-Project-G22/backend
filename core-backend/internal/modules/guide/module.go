package guide

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/domain/port"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/infrastructure/repository"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	iammiddleware "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"guide",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	fx.Provide(
		fx.Annotate(
			infrarepo.NewGuideRepository,
			fx.As(new(repository.GuideRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewStepRepository,
			fx.As(new(repository.StepRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewProgressRepository,
			fx.As(new(repository.ProgressRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			appusecase.NewGuideViewUsecase,
			fx.As(new(usecase.GuideViewUseCase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewGuideAdminUsecase,
			fx.As(new(usecase.GuideManagementUseCase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewJourneyManagementUsecase,
			fx.As(new(usecase.JourneyManagementUseCase)),
		),
	),

	// AI tool handlers
	fx.Provide(
		fx.Annotate(
			NewSearchGuidesTool,
			fx.As(new(port.ToolHandler)),
			fx.ResultTags(`group:"ai_tool_handlers"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewGuideDetailTool,
			fx.As(new(port.ToolHandler)),
			fx.ResultTags(`group:"ai_tool_handlers"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewGuideProgressTool,
			fx.As(new(port.ToolHandler)),
			fx.ResultTags(`group:"ai_tool_handlers"`),
		),
	),

	fx.Provide(handler.NewGuideViewHandler),
	fx.Provide(handler.NewGuideAdminHandler),

	fx.Provide(
		fx.Annotate(
			GuideSeedPermissions,
			fx.ResultTags(`group:"permission_seeds"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			GuideSeedRoles,
			fx.ResultTags(`group:"role_seeds"`),
		),
	),

	fx.Invoke(func(api huma.API, guideViewHandler *handler.GuideViewHandler, guideAdminHandler *handler.GuideAdminHandler, tokenService token.TokenService, authService iamservice.AuthService, roleAssignmentUsecase iamusecase.RoleAssignmentUsecase) {
		authMiddleware := iammiddleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := iammiddleware.AccountStatusMiddleware(api, authService)
		readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, GuideRead, []string{"super_admin"})
		writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, GuideWrite, []string{"super_admin"})
		updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, GuideUpdate, []string{"super_admin"})
		deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, GuideDelete, []string{"super_admin"})
		routes.RegisterRoutes(api, routes.RouteDependencies{
			GuideViewHandler:           guideViewHandler,
			GuideAdminHandler:          guideAdminHandler,
			AuthMiddleware:             authMiddleware,
			AccountStatusMiddleware:    accountStatusMiddleware,
			ReadPermissionMiddleware:   readPermissionMiddleware,
			WritePermissionMiddleware:  writePermissionMiddleware,
			UpdatePermissionMiddleware: updatePermissionMiddleware,
			DeletePermissionMiddleware: deletePermissionMiddleware,
		})
	}),
)
