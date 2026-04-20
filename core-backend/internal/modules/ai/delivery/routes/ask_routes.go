package routes

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

const (
	aiBase = "/api/v1/ai"
)

func registerAskRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "ask",
		Method:      "POST",
		Path:        aiBase + "/ask",
		Summary:     "Ask a question",
		Description: "Ask AI a question and get an answer with citations.",
		Tags:        []string{"AI Ask"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AskHandler.HandleAsk)

	huma.Register(api, huma.Operation{
		OperationID: "listConversations",
		Method:      "GET",
		Path:        aiBase + "/conversations",
		Summary:     "List conversations",
		Description: "Get a paginated list of user's conversation sessions.",
		Tags:        []string{"AI Conversations"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.ListConversationsInput) (*dto.ListConversationsOutput, error) {
		accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
		userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
		return deps.AskHandler.HandleListConversations(ctx, input, accountID, userID)
	})

	huma.Register(api, huma.Operation{
		OperationID: "getConversation",
		Method:      "GET",
		Path:        aiBase + "/conversations/:sessionId",
		Summary:     "Get conversation",
		Description: "Get a conversation with its messages.",
		Tags:        []string{"AI Conversations"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.GetConversationInput) (*dto.GetConversationOutput, error) {
		sessionID := getSessionIDFromContext(ctx)
		accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
		return deps.AskHandler.HandleGetConversation(ctx, input, sessionID, accountID)
	})

	huma.Register(api, huma.Operation{
		OperationID: "archiveConversation",
		Method:      "DELETE",
		Path:        aiBase + "/conversations/:sessionId",
		Summary:     "Archive conversation",
		Description: "Archive (soft-delete) a conversation session.",
		Tags:        []string{"AI Conversations"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.ArchiveConversationInput) (*dto.ArchiveConversationOutput, error) {
		sessionID := getSessionIDFromContext(ctx)
		accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
		return deps.AskHandler.HandleArchiveConversation(ctx, input, sessionID, accountID)
	})
}

func getSessionIDFromContext(ctx context.Context) uuid.UUID {
	val := ctx.Value("sessionId")
	if s, ok := val.(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			return id
		}
	}
	return uuid.Nil
}
