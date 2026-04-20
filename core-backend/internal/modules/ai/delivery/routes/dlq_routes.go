package routes

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/danielgtaylor/huma/v2"
)

const (
	aiIngestionDLQBase = aiIngestionBase + "/dlq"
)

func registerDLQRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listDeadEvents",
		Method:      "GET",
		Path:        aiIngestionDLQBase + "/events",
		Summary:     "List dead letter events",
		Description: "Get a paginated list of dead letter queue events for the account.",
		Tags:        []string{"AI DLQ"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.DLQHandler.HandleListDeadEvents)

	huma.Register(api, huma.Operation{
		OperationID: "getDeadEvent",
		Method:      "GET",
		Path:        aiIngestionDLQBase + "/events/{eventId}",
		Summary:     "Get dead event",
		Description: "Get a specific dead letter queue event by ID.",
		Tags:        []string{"AI DLQ"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.DLQHandler.HandleGetDeadEvent)

	huma.Register(api, huma.Operation{
		OperationID: "redriveEvent",
		Method:      "POST",
		Path:        aiIngestionDLQBase + "/events/{eventId}/redrive",
		Summary:     "Redrive single event",
		Description: "Redrive a dead event back to the processing queue.",
		Tags:        []string{"AI DLQ"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.DLQHandler.HandleRedriveEvent)

	huma.Register(api, huma.Operation{
		OperationID: "redriveBatch",
		Method:      "POST",
		Path:        aiIngestionDLQBase + "/events/batch/redrive",
		Summary:     "Redrive batch",
		Description: "Redrive multiple dead events in a single request.",
		Tags:        []string{"AI DLQ"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.DLQHandler.HandleRedriveBatch)
}

func registerIngestionStatusStreamRoute(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "streamIngestionStatus",
		Method:      "GET",
		Path:        aiIngestionStatusBase + "/stream",
		Summary:     "Stream ingestion status",
		Description: "Subscribe to real-time ingestion status updates via SSE.",
		Tags:        []string{"AI Ingestion Status"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.StreamStatusInput) (*huma.StreamResponse, error) {
		return &huma.StreamResponse{
			Body: func(hctx huma.Context) {
				writer := service.NewSSEWriter(hctx)
				writer.WriteHeaders()
				lastEventID := hctx.Header("Last-Event-ID")
				err := deps.SSEHandler.StreamAccountStatus(ctx, lastEventID, func(event string, payload any) error {
					return writer.WriteEvent(event, payload)
				})
				if err != nil && ctx.Err() == nil {
					_ = writer.WriteEvent("error", map[string]string{"message": err.Error()})
				}
			},
		}, nil
	})
}

func registerIngestToggleRoute(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "getIngestionToggle",
		Method:      "GET",
		Path:        aiIngestionBase + "/toggle",
		Summary:     "Get ingestion toggle",
		Description: "Get the current ingestion toggle state.",
		Tags:        []string{"AI Ingestion"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.ToggleHandler.HandleGetIngestToggle)

	huma.Register(api, huma.Operation{
		OperationID: "setIngestionToggle",
		Method:      "PATCH",
		Path:        aiIngestionBase + "/toggle",
		Summary:     "Set ingestion toggle",
		Description: "Enable or disable document ingestion for the account.",
		Tags:        []string{"AI Ingestion"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.ToggleHandler.HandleSetIngestToggle)
}
