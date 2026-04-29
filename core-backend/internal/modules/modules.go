package modules

import (
	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	iamnotification "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/infrastructure/notification"
	appusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/usecase"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"go.uber.org/fx"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/coregrpc"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification"
)

var Modules = fx.Options(
	ai.Module,
	iam.Module,
	guide.Module,
	community.Module,
	coregrpc.Module,
	notification.Module,
	library.Module,

	// Override notification's default IAM readers with real IAM-backed adapters.
	fx.Decorate(func(
		_ appusecase.IAMGlobalPreferenceReader,
		prefRepo iamrepo.NotificationPreferenceRepository,
	) appusecase.IAMGlobalPreferenceReader {
		return iamnotification.NewIAMGlobalPreferenceReaderAdapter(prefRepo)
	}),

	fx.Decorate(func(
		_ notifrepo.AccountReader,
		accountRepo iamrepo.AccountRepository,
	) notifrepo.AccountReader {
		return iamnotification.NewAccountReaderAdapter(accountRepo)
	}),
)
