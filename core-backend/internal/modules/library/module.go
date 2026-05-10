package library

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	iammiddleware "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/delivery/routes"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/infrastructure/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/infrastructure/tier"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"library",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),

	fx.Provide(
		fx.Annotate(
			infrarepo.NewLibraryCategoryRepository,
			fx.As(new(repository.LibraryCategoryRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewLibraryTemplateGroupRepository,
			fx.As(new(repository.LibraryTemplateGroupRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewLibraryTemplateRepository,
			fx.As(new(repository.LibraryTemplateRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewLibraryInteractiveFormRepository,
			fx.As(new(repository.LibraryInteractiveFormRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			infrarepo.NewLibraryTemplateDownloadRepository,
			fx.As(new(repository.LibraryTemplateDownloadRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			appusecase.NewLibraryViewUsecase,
			fx.As(new(usecase.LibraryViewUsecase)),
		),
	),
	fx.Provide(
		fx.Annotate(
			appusecase.NewLibraryAdminUsecase,
			fx.As(new(usecase.LibraryAdminUsecase)),
		),
	),

	fx.Provide(
		fx.Annotate(
			service.NewTemplateFileValidator,
		),
	),
	fx.Provide(
		fx.Annotate(
			service.NewLibraryService,
			fx.As(new(service.LibraryService)),
		),
	),

	fx.Provide(
		fx.Annotate(
			tier.NewTierServiceStub,
			fx.As(new(usecase.TierService)),
		),
	),

	fx.Provide(handler.NewLibraryHandler),
	fx.Provide(handler.NewLibraryAdminHandler),

	fx.Invoke(func(
		api huma.API,
		libraryHandler *handler.LibraryHandler,
		libraryAdminHandler *handler.LibraryAdminHandler,
		tokenService token.TokenService,
		authService iamservice.AuthService,
		roleAssignmentUsecase iamusecase.RoleAssignmentUsecase,
	) {
		authMiddleware := iammiddleware.AuthMiddleware(api, tokenService, authService)
		accountStatusMiddleware := iammiddleware.AccountStatusMiddleware(api, authService)
		readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.IAMRead, nil)
		writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.IAMWrite, nil)
		updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.IAMUpdate, nil)
		deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, roleAssignmentUsecase, permissions.IAMDelete, nil)

		routes.RegisterRoutes(api, routes.RouteDependencies{
			ViewHandler:                libraryHandler,
			AdminHandler:               libraryAdminHandler,
			AuthMiddleware:             authMiddleware,
			AccountStatusMiddleware:    accountStatusMiddleware,
			ReadPermissionMiddleware:   readPermissionMiddleware,
			WritePermissionMiddleware:  writePermissionMiddleware,
			UpdatePermissionMiddleware: updatePermissionMiddleware,
			DeletePermissionMiddleware: deletePermissionMiddleware,
		})
	}),
)
