package ai_tool

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/domain/port"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/domain/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Module("ai-tool",
	fx.Provide(service.NewToolRegistry),
	fx.Provide(server.NewAIToolService),

	fx.Invoke(
		fx.Annotate(
			registerHandlers,
			fx.ParamTags(``, `group:"ai_tool_handlers"`),
		),
	),
)

func registerHandlers(registry *service.ToolRegistry, handlers []port.ToolHandler) {
	for _, h := range handlers {
		registry.Register(h)
	}
}
