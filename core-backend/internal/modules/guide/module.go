package guide

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/application/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	infrarepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/infrastructure/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
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
			func(db *core.Database) sharedrepo.Transactor { return db },
			fx.As(new(sharedrepo.Transactor)),
		),
	),

	fx.Provide(
		fx.Annotate(
			infrarepo.NewCategoryRepository,
			fx.As(new(repository.CategoryRepository)),
		),
	),
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
)
