package community

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/infrastructure/repository"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	iammiddleware "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/taxonomy"
	"github.com/Final-Year-Project-G22/backend/core/internal/ws"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"community",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	fx.Provide(
		fx.Annotate(
			infrarepo.NewCommunityCategoryRepository,
			fx.As(new(repository.CommunityCategoryRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewDiscussionThreadRepository,
			fx.As(new(repository.DiscussionThreadRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewDiscussionPostRepository,
			fx.As(new(repository.DiscussionPostRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewUserThreadSettingsRepository,
			fx.As(new(repository.UserThreadSettingsRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewUserCategorySettingsRepository,
			fx.As(new(repository.UserCategorySettingsRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewAttachmentRepository,
			fx.As(new(repository.AttachmentRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewContentReportRepository,
			fx.As(new(repository.ContentReportRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewThreadBlockedUserRepository,
			fx.As(new(repository.ThreadBlockedUserRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewCommunityCategoryUsecase,
			fx.As(new(usecase.CommunityCategoryUsecase)),
		),
	),
	fx.Provide(taxonomy.NewTaxonomyValidator),
	fx.Provide(
		fx.Annotate(
			appusecase.NewDiscussionThreadUsecase,
			fx.As(new(usecase.DiscussionThreadUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewDiscussionPostUsecase,
			fx.As(new(usecase.DiscussionPostUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewCommunityFollowUsecase,
			fx.As(new(usecase.CommunityFollowUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewContentReportUsecase,
			fx.As(new(usecase.ContentReportUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewThreadBlockUsecase,
			fx.As(new(usecase.ThreadBlockUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			service.NewAttachmentService,
			fx.As(new(usecase.AttachmentUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			service.NewCommunityService,
			fx.As(new(service.CommunityService)),
		),
	),
	fx.Provide(
		fx.Annotate(
			service.NewCommunityAttachmentValidator,
			fx.As(new(service.AttachmentValidator)),
		),
	),
	fx.Provide(service.NewAttachmentCleanupWorker),
	fx.Provide(handler.NewCommunityAdminHandler),
	fx.Provide(service.NewCommunityAttachmentValidator),
	fx.Provide(handler.NewCommunityHandler),
	fx.Provide(
		fx.Annotate(
			CommunitySeedPermissions,
			fx.ResultTags(`group:"permission_seeds"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			CommunitySeedRoles,
			fx.ResultTags(`group:"role_seeds"`),
		),
	),
	fx.Provide(ws.NewHub),
	fx.Provide(ws.NewHandler),
	fx.Invoke(func(
		api huma.API,
		engine *gin.Engine,
		communityHandler *handler.CommunityHandler,
		communityAdminHandler *handler.CommunityAdminHandler,
		attachmentCleanupWorker *service.AttachmentCleanupWorker,
		wsHandler *ws.Handler,
		tokenService token.TokenService,
		authService iamservice.AuthService,
		roleAssignmentUsecase iamusecase.RoleAssignmentUsecase,
	) {
		engine.GET("/ws", gin.WrapH(wsHandler))
		authMiddleware := iammiddleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := iammiddleware.AccountStatusMiddleware(api, authService)
		readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, CommunityRead, []string{"super_admin"})
		writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, CommunityWrite, []string{"super_admin"})
		updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, CommunityUpdate, []string{"super_admin"})
		deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, CommunityDelete, []string{"super_admin"})

		routes.RegisterRoutes(api, routes.RouteDependencies{
			CommunityHandler:           communityHandler,
			CommunityAdminHandler:      communityAdminHandler,
			AuthMiddleware:             authMiddleware,
			AccountStatusMiddleware:    accountStatusMiddleware,
			ReadPermissionMiddleware:   readPermissionMiddleware,
			WritePermissionMiddleware:  writePermissionMiddleware,
			UpdatePermissionMiddleware: updatePermissionMiddleware,
			DeletePermissionMiddleware: deletePermissionMiddleware,
		})

		attachmentCleanupWorker.Start(context.Background())
	}),
)
