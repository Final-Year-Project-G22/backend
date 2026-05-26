package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	pb "github.com/Final-Year-Project-G22/backend/core/pb/ai/conversation/v1"
	pb_inference "github.com/Final-Year-Project-G22/backend/core/pb/ai/inference/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InferenceGRPCClient struct {
	client             pb_inference.AIInferenceServiceClient
	conversationClient pb.AIConversationServiceClient
	timeout            time.Duration
	authToken          string
	logger             core.Logger
}

func NewInferenceGRPCClient(cfg *core.Config, logger core.Logger) (port.AIInferencePort, error) {
	endpoint := cfg.AI.InferenceGRPCEndpoint
	if endpoint == "" {
		return nil, errors.New("ai inference grpc endpoint is required")
	}

	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create ai inference grpc client: %w", err)
	}

	return &InferenceGRPCClient{
		client:             pb_inference.NewAIInferenceServiceClient(conn),
		conversationClient: pb.NewAIConversationServiceClient(conn),
		timeout:            cfg.AI.InferenceTimeout,
		authToken:          cfg.AI.InferenceAuthToken,
		logger:             logger,
	}, nil
}

func (c *InferenceGRPCClient) Ask(ctx context.Context, req port.AskRequest) (port.AskResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if req.TopK < 1 || req.TopK > 20 {
		return port.AskResponse{}, fmt.Errorf("top_k must be between 1 and 20, got %d", req.TopK)
	}

	if c.authToken != "" {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "authorization", "Bearer "+c.authToken)
	}

	protoReq := &pb_inference.AskRequest{
		RequestId: req.RequestID.String(),
		UserId:    req.UserID.String(),
		AccountId: req.AccountID.String(),
		Query:     req.Query,
		Language:  string(req.Language),
		TopK:      req.TopK,
		Strategy:  req.Strategy,
		DebugMode: req.DebugMode,
	}
	if req.SessionID != nil {
		sid := req.SessionID.String()
		protoReq.SessionId = &sid
	}
	if req.Title != nil {
		protoReq.Title = req.Title
	}

	// Pass taxonomy filters as gRPC metadata for retrieval narrowing.
	if len(req.SectorIDs) > 0 {
		var parts []string
		for _, id := range req.SectorIDs {
			parts = append(parts, id.String())
		}
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-sector-ids", strings.Join(parts, ","))
	}
	if len(req.TagIDs) > 0 {
		var parts []string
		for _, id := range req.TagIDs {
			parts = append(parts, id.String())
		}
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-tag-ids", strings.Join(parts, ","))
	}
	if req.Region != nil {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-region", *req.Region)
	}
	if req.Stage != nil {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-stage", *req.Stage)
	}

	resp, err := c.client.Ask(callCtx, protoReq)
	if err != nil {
		return port.AskResponse{}, mapAskError(err)
	}

	requestID, err := uuid.Parse(resp.GetRequestId())
	if err != nil {
		return port.AskResponse{}, fmt.Errorf("parse ask response request_id: %w", err)
	}

	sessionID, err := uuid.Parse(resp.GetSessionId())
	if err != nil {
		return port.AskResponse{}, fmt.Errorf("parse ask response session_id: %w", err)
	}

	citations := make([]port.Citation, 0, len(resp.GetCitations()))
	for _, citation := range resp.GetCitations() {
		documentID, parseErr := uuid.Parse(citation.GetDocumentId())
		if parseErr != nil {
			return port.AskResponse{}, fmt.Errorf("parse citation document_id: %w", parseErr)
		}

		chunkID, parseErr := uuid.Parse(citation.GetChunkId())
		if parseErr != nil {
			return port.AskResponse{}, fmt.Errorf("parse citation chunk_id: %w", parseErr)
		}

		var title *string
		if citation.Title != nil {
			t := citation.GetTitle()
			title = &t
		}
		var excerpt *string
		if citation.Excerpt != nil {
			e := citation.GetExcerpt()
			excerpt = &e
		}

		citations = append(citations, port.Citation{
			DocumentID: documentID,
			ChunkID:    chunkID,
			SourceType: citation.GetSourceType(),
			Title:      title,
			Score:      citation.GetScore(),
			Excerpt:    excerpt,
		})
	}

	usage := resp.GetUsage()
	usageData := port.Usage{}
	if usage != nil {
		usageData = port.Usage{
			PromptTokens:     int(usage.GetPromptTokens()),
			CompletionTokens: int(usage.GetCompletionTokens()),
			TotalTokens:      int(usage.GetTotalTokens()),
		}
	}

	return port.AskResponse{
		RequestID: requestID,
		SessionID: sessionID,
		CreatedAt: resp.GetSessionCreatedAt(),
		UpdatedAt: resp.GetSessionUpdatedAt(),
		Answer:    resp.GetAnswer(),
		Citations: citations,
		Usage:     usageData,
		Model:     resp.GetModel(),
		LatencyMS: int(resp.GetLatencyMs()),
	}, nil
}

func mapAskError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("ai inference request failed: %w", err)
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return fmt.Errorf("ai inference invalid argument: %w", err)
	case codes.Unauthenticated:
		return fmt.Errorf("ai inference authentication failed: %w", err)
	case codes.ResourceExhausted:
		return fmt.Errorf("ai inference quota exceeded: %w", err)
	case codes.DeadlineExceeded:
		return fmt.Errorf("ai inference timeout: %w", err)
	default:
		return fmt.Errorf("ai inference request failed with %s: %w", st.Code(), err)
	}
}

func mapAskStreamError(err error) *port.ErrorInfo {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return &port.ErrorInfo{Code: "INTERNAL", Message: err.Error()}
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return &port.ErrorInfo{Code: "INVALID_ARGUMENT", Message: st.Message()}
	case codes.Unauthenticated:
		return &port.ErrorInfo{Code: "UNAUTHENTICATED", Message: st.Message()}
	case codes.ResourceExhausted:
		return &port.ErrorInfo{Code: "RESOURCE_EXHAUSTED", Message: st.Message()}
	case codes.DeadlineExceeded:
		return &port.ErrorInfo{Code: "DEADLINE_EXCEEDED", Message: st.Message()}
	default:
		return &port.ErrorInfo{Code: st.Code().String(), Message: st.Message()}
	}
}

func (c *InferenceGRPCClient) AskStream(ctx context.Context, req port.AskRequest) (<-chan port.AskStreamChunk, error) {
	if req.TopK < 1 || req.TopK > 20 {
		return nil, fmt.Errorf("top_k must be between 1 and 20, got %d", req.TopK)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)

	if c.authToken != "" {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "authorization", "Bearer "+c.authToken)
	}

	protoReq := &pb_inference.AskRequest{
		RequestId: req.RequestID.String(),
		UserId:    req.UserID.String(),
		AccountId: req.AccountID.String(),
		Query:     req.Query,
		Language:  string(req.Language),
		TopK:      req.TopK,
		Strategy:  req.Strategy,
		DebugMode: req.DebugMode,
	}
	if req.SessionID != nil {
		sid := req.SessionID.String()
		protoReq.SessionId = &sid
	}
	if req.Title != nil {
		protoReq.Title = req.Title
	}

	// Pass taxonomy filters as gRPC metadata for retrieval narrowing.
	if len(req.SectorIDs) > 0 {
		var parts []string
		for _, id := range req.SectorIDs {
			parts = append(parts, id.String())
		}
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-sector-ids", strings.Join(parts, ","))
	}
	if len(req.TagIDs) > 0 {
		var parts []string
		for _, id := range req.TagIDs {
			parts = append(parts, id.String())
		}
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-tag-ids", strings.Join(parts, ","))
	}
	if req.Region != nil {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-region", *req.Region)
	}
	if req.Stage != nil {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-taxonomy-stage", *req.Stage)
	}

	stream, err := c.client.AskStream(callCtx, protoReq)
	if err != nil {
		cancel()
		return nil, mapAskError(err)
	}

	ch := make(chan port.AskStreamChunk, 10)
	go func() {
		defer close(ch)
		defer cancel()
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) || ctx.Err() != nil {
					return
				}
				if c.logger != nil {
					c.logger.Warn("ask stream recv failed", core.Error(err))
				}
				info := mapAskStreamError(err)
				if info != nil {
					ch <- port.AskStreamChunk{Error: info}
				}
				return
			}

			out := port.AskStreamChunk{}
			if text := chunk.GetText(); text != nil {
				t := text.GetText()
				out.Text = &t
			}
			if done := chunk.GetDone(); done != nil {
				sessionID, _ := uuid.Parse(done.GetSessionId())
				out.Done = &port.DoneInfo{
					Model:     done.GetModel(),
					Usage:     mapUsage(done.GetUsage()),
					LatencyMs: int(done.GetLatencyMs()),
					SessionID: sessionID,
					CreatedAt: done.GetSessionCreatedAt(),
					UpdatedAt: done.GetSessionUpdatedAt(),
				}
			}
			if errChunk := chunk.GetError(); errChunk != nil {
				out.Error = &port.ErrorInfo{
					Code:    errChunk.GetCode(),
					Message: errChunk.GetMessage(),
				}
			}
			if toolUse := chunk.GetToolUse(); toolUse != nil {
				out.ToolUse = &port.ToolUseInfo{
					Tool:          toolUse.GetTool(),
					ArgumentsJSON: toolUse.GetArgumentsJson(),
				}
			}
			if toolResult := chunk.GetToolResult(); toolResult != nil {
				out.ToolResult = &port.ToolResultInfo{
					Tool:          toolResult.GetTool(),
					ResultSummary: toolResult.GetResultSummary(),
				}
			}
			if thinking := chunk.GetThinking(); thinking != nil {
				t := thinking.GetText()
				out.Thinking = &t
			}
			if citations := chunk.GetCitations(); citations != nil {
				out.Citations = make([]port.Citation, 0, len(citations.GetCitations()))
				for _, cit := range citations.GetCitations() {
					docID, _ := uuid.Parse(cit.GetDocumentId())
					chunkID, _ := uuid.Parse(cit.GetChunkId())
					var title *string
					if cit.Title != nil {
						t := cit.GetTitle()
						title = &t
					}
					var excerpt *string
					if cit.Excerpt != nil {
						e := cit.GetExcerpt()
						excerpt = &e
					}
					out.Citations = append(out.Citations, port.Citation{
						DocumentID: docID,
						ChunkID:    chunkID,
						SourceType: cit.GetSourceType(),
						Title:      title,
						Score:      cit.GetScore(),
						Excerpt:    excerpt,
					})
				}
			}
			ch <- out
		}
	}()

	return ch, nil
}

func mapUsage(u *pb_inference.Usage) port.Usage {
	if u == nil {
		return port.Usage{}
	}
	return port.Usage{
		PromptTokens:     int(u.GetPromptTokens()),
		CompletionTokens: int(u.GetCompletionTokens()),
		TotalTokens:      int(u.GetTotalTokens()),
	}
}

func (c *InferenceGRPCClient) ListConversations(ctx context.Context, req port.ListConversationsRequest) (port.ListConversationsResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.authToken != "" {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "authorization", "Bearer "+c.authToken)
	}

	protoReq := &pb.ListConversationsRequest{
		UserId:    req.UserID.String(),
		AccountId: req.AccountID.String(),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	resp, err := c.conversationClient.ListConversations(callCtx, protoReq)
	if err != nil {
		return port.ListConversationsResponse{}, mapConversationError(err)
	}

	sessions := make([]port.Conversation, 0, len(resp.GetSessions()))
	for _, s := range resp.GetSessions() {
		id, _ := uuid.Parse(s.GetId())
		accID, _ := uuid.Parse(s.GetAccountId())
		sessions = append(sessions, port.Conversation{
			ID:        id,
			AccountID: accID,
			Title:     s.GetTitle(),
			Language:  constants.Locale(s.GetLanguage()),
			CreatedAt: s.GetCreatedAt(),
			UpdatedAt: s.GetUpdatedAt(),
		})
	}

	return port.ListConversationsResponse{
		Sessions: sessions,
		Total:    resp.GetTotal(),
	}, nil
}

func (c *InferenceGRPCClient) GetConversation(ctx context.Context, req port.GetConversationRequest) (port.GetConversationResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.authToken != "" {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "authorization", "Bearer "+c.authToken)
	}

	protoReq := &pb.GetConversationRequest{
		SessionId:      req.SessionID.String(),
		AccountId:      req.AccountID.String(),
		MessageLimit:   req.MessageLimit,
		MessageOffset:  req.MessageOffset,
		IncludeDeleted: req.IncludeDeleted,
	}

	resp, err := c.conversationClient.GetConversation(callCtx, protoReq)
	if err != nil {
		return port.GetConversationResponse{}, mapConversationError(err)
	}

	s := resp.GetSession()
	sessID, _ := uuid.Parse(s.GetId())
	accID, _ := uuid.Parse(s.GetAccountId())

	messages := make([]port.Message, 0, len(resp.GetMessages()))
	for _, m := range resp.GetMessages() {
		messages = append(messages, port.Message{
			ID:         uuid.MustParse(m.GetId()),
			Role:       m.GetRole(),
			Content:    m.GetContent(),
			Citations:  mapCitations(m.GetCitations()),
			TokenUsage: mapConversationUsage(m.GetUsage()),
			CreatedAt:  m.GetCreatedAt(),
		})
	}

	return port.GetConversationResponse{
		Session: port.Conversation{
			ID:        sessID,
			AccountID: accID,
			Title:     s.GetTitle(),
			Language:  constants.Locale(s.GetLanguage()),
			CreatedAt: s.GetCreatedAt(),
			UpdatedAt: s.GetUpdatedAt(),
		},
		Messages:  messages,
		TotalMsgs: resp.GetTotalMessages(),
	}, nil
}

func (c *InferenceGRPCClient) ArchiveConversation(ctx context.Context, req port.ArchiveConversationRequest) (port.ArchiveConversationResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.authToken != "" {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "authorization", "Bearer "+c.authToken)
	}

	protoReq := &pb.ArchiveConversationRequest{
		SessionId: req.SessionID.String(),
		AccountId: req.AccountID.String(),
	}

	resp, err := c.conversationClient.ArchiveConversation(callCtx, protoReq)
	if err != nil {
		return port.ArchiveConversationResponse{}, mapConversationError(err)
	}

	return port.ArchiveConversationResponse{Success: resp.GetSuccess(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

func mapConversationError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("ai conversation request failed: %w", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("session not found: %w", err)
	case codes.InvalidArgument:
		return fmt.Errorf("invalid argument: %w", err)
	case codes.Unauthenticated:
		return fmt.Errorf("authentication failed: %w", err)
	default:
		return fmt.Errorf("conversation request failed with %s: %w", st.Code(), err)
	}
}

func mapConversationUsage(u *pb.Usage) port.Usage {
	if u == nil {
		return port.Usage{}
	}
	return port.Usage{
		PromptTokens:     int(u.GetPromptTokens()),
		CompletionTokens: int(u.GetCompletionTokens()),
		TotalTokens:      int(u.GetTotalTokens()),
	}
}

func mapCitations(citations []*pb.Citation) []port.Citation {
	if citations == nil {
		return nil
	}
	result := make([]port.Citation, 0, len(citations))
	for _, c := range citations {
		docID, _ := uuid.Parse(c.GetDocumentId())
		chunkID, _ := uuid.Parse(c.GetChunkId())
		result = append(result, port.Citation{
			DocumentID: docID,
			ChunkID:    chunkID,
			SourceType: c.GetSourceType(),
			Title:      nil,
			Score:      c.GetScore(),
		})
	}
	return result
}
