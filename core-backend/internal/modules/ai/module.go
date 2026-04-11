package ai

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	aiinfraclient "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/infrastructure/client"
	"go.uber.org/fx"
)

var Module = fx.Module("ai",
	fx.Provide(
		fx.Annotate(
			aiinfraclient.NewInferenceGRPCClient,
			fx.As(new(port.AIInferencePort)),
		),
	),
)
