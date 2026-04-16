package community

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/service"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/infrastructure/repository"
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
			service.NewCommunityService,
			fx.As(new(service.CommunityService)),
		),
	),
	fx.Provide(service.NewCommunityAttachmentValidator),
)
