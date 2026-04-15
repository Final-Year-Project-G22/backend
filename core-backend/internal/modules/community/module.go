package community

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
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
)
