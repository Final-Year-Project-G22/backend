package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	pb "github.com/Final-Year-Project-G22/backend/core/pb/ai/inference/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InferenceGRPCClient struct {
	client    pb.AIInferenceServiceClient
	timeout   time.Duration
	authToken string
}

func NewInferenceGRPCClient(cfg *core.Config) (port.AIInferencePort, error) {
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
		client:    pb.NewAIInferenceServiceClient(conn),
		timeout:   cfg.AI.InferenceTimeout,
		authToken: cfg.AI.InferenceAuthToken,
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

	protoReq := &pb.AskRequest{
		RequestId: req.RequestID.String(),
		UserId:    req.UserID.String(),
		AccountId: req.AccountID.String(),
		Query:     req.Query,
		Language:  string(req.Language),
		TopK:      req.TopK,
	}
	if req.SessionID != nil {
		sid := req.SessionID.String()
		protoReq.SessionId = &sid
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

		citations = append(citations, port.Citation{
			DocumentID: documentID,
			ChunkID:    chunkID,
			SourceType: citation.GetSourceType(),
			Title:      title,
			Score:      citation.GetScore(),
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
