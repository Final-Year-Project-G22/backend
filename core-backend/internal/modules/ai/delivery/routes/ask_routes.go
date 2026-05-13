package routes

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
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
		OperationID: "askStream",
		Method:      "POST",
		Path:        aiBase + "/ask/stream",
		Summary:     "Ask a question with streaming response",
		Description: "Ask AI a question and receive token chunks as Server-Sent Events.",
		Tags:        []string{"AI Ask"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, input *dto.AskStreamInput) (*huma.StreamResponse, error) {
		chunks, err := deps.AskHandler.HandleAskStream(ctx, input)
		if err != nil {
			return nil, err
		}

		return &huma.StreamResponse{
			Body: func(hctx huma.Context) {
				writer := service.NewSSEWriter(hctx)
				writer.WriteHeaders()
				_ = writer.WriteComment("ask-stream-started")

				for {
					select {
					case <-ctx.Done():
						_ = writer.WriteComment("ask-stream-cancelled")
						return
					case chunk, ok := <-chunks:
						if !ok {
							return
						}

						if chunk.Text != nil {
							_ = writer.WriteEvent("chunk", dto.AskStreamChunkEventBody{
								Text: *chunk.Text,
							})
						}

						if len(chunk.Citations) > 0 {
							citations := make([]dto.CitationDTO, 0, len(chunk.Citations))
							for _, c := range chunk.Citations {
								citations = append(citations, dto.CitationDTO{
									DocumentID: c.DocumentID,
									ChunkID:    c.ChunkID,
									SourceType: c.SourceType,
									Title:      c.Title,
									Score:      c.Score,
									Excerpt:    c.Excerpt,
								})
							}

							_ = writer.WriteEvent("citations", dto.AskStreamCitationEventBody{
								Citations: citations,
							})
						}

						if chunk.Done != nil {
							_ = writer.WriteEvent("done", dto.AskStreamDoneEventBody{
								Model:     chunk.Done.Model,
								LatencyMS: chunk.Done.LatencyMs,
								Usage: dto.UsageDTO{
									PromptTokens:     chunk.Done.Usage.PromptTokens,
									CompletionTokens: chunk.Done.Usage.CompletionTokens,
									TotalTokens:      chunk.Done.Usage.TotalTokens,
								},
								SessionID: chunk.Done.SessionID,
								CreatedAt: parseTimeFromRFC3339(chunk.Done.CreatedAt),
								UpdatedAt: parseTimeFromRFC3339(chunk.Done.UpdatedAt),
							})
							return
						}

						if chunk.Error != nil {
							_ = writer.WriteEvent("error", dto.AskStreamErrorEventBody{
								Code:    chunk.Error.Code,
								Message: chunk.Error.Message,
							})
							return
						}
					}
				}
			},
		}, nil
	})

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
		Path:        aiBase + "/conversations/{sessionId}",
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
		Path:        aiBase + "/conversations/{sessionId}",
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
	val := ctx.Value(contextkeys.SessionID)
	if s, ok := val.(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func parseTimeFromRFC3339(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return t
	}
	t, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return t
	}
	return time.Time{}
}
