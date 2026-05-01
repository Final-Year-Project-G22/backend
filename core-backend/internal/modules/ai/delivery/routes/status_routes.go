package routes

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

const (
	aiIngestionStatusBase = "/api/v1/ai/ingestion/status"
)

func registerStatusRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "getIngestionStatusByDocumentID",
		Method:      "GET",
		Path:        aiIngestionStatusBase + "/documents/{documentId}",
		Summary:     "Get ingestion status by document ID",
		Description: "Returns the current ingestion status for a document.",
		Tags:        []string{"AI Ingestion Status"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.GetStatusByDocumentInput) (*dto.GetStatusByDocumentOutput, error) {
		docIDVal := ctx.Value(contextkeys.DocumentID)
		if docIDVal == nil {
			return nil, nil
		}
		docIDStr, ok := docIDVal.(string)
		if !ok {
			return nil, nil
		}
		docID, err := uuid.Parse(docIDStr)
		if err != nil {
			return nil, err
		}
		return deps.StatusHandler.HandleGetStatusByDocumentID(ctx, docID)
	})

	huma.Register(api, huma.Operation{
		OperationID: "listIngestionStatusByAccountID",
		Method:      "GET",
		Path:        aiIngestionStatusBase + "/accounts/{accountId}",
		Summary:     "List ingestion status by account ID",
		Description: "Returns a paginated list of ingestion statuses for an account.",
		Tags:        []string{"AI Ingestion Status"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.ListStatusByAccountInput) (*dto.ListStatusByAccountOutput, error) {
		limit := 20
		offset := 0
		if input.Query.Limit != nil {
			limit = *input.Query.Limit
		}
		if input.Query.Offset != nil {
			offset = *input.Query.Offset
		}
		return deps.StatusHandler.HandleListStatusByAccountID(ctx, input.AccountID, limit, offset)
	})

	huma.Register(api, huma.Operation{
		OperationID: "listIngestionStatusByUserID",
		Method:      "GET",
		Path:        aiIngestionStatusBase + "/users/{userId}",
		Summary:     "List ingestion status by user ID",
		Description: "Returns a paginated list of ingestion statuses for a user.",
		Tags:        []string{"AI Ingestion Status"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.ListStatusByUserInput) (*dto.ListStatusByUserOutput, error) {
		limit := 20
		offset := 0
		if input.Query.Limit != nil {
			limit = *input.Query.Limit
		}
		if input.Query.Offset != nil {
			offset = *input.Query.Offset
		}
		return deps.StatusHandler.HandleListStatusByUserID(ctx, input.UserID, limit, offset)
	})
}
