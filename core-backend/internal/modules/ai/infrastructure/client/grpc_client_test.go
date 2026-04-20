package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	pb "github.com/Final-Year-Project-G22/backend/core/pb/ai/conversation/v1"
	pb_inference "github.com/Final-Year-Project-G22/backend/core/pb/ai/inference/v1"
	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeInferenceClient struct {
	askFn       func(ctx context.Context, in *pb_inference.AskRequest, opts ...grpclib.CallOption) (*pb_inference.AskResponse, error)
	askStreamFn func(ctx context.Context, in *pb_inference.AskRequest, opts ...grpclib.CallOption) (grpclib.ServerStreamingClient[pb_inference.AskStreamChunk], error)
}

func (f *fakeInferenceClient) Ask(
	ctx context.Context,
	in *pb_inference.AskRequest,
	opts ...grpclib.CallOption,
) (*pb_inference.AskResponse, error) {
	if f.askFn == nil {
		return nil, errors.New("askFn not set")
	}
	return f.askFn(ctx, in, opts...)
}

func (f *fakeInferenceClient) AskStream(
	ctx context.Context,
	in *pb_inference.AskRequest,
	opts ...grpclib.CallOption,
) (grpclib.ServerStreamingClient[pb_inference.AskStreamChunk], error) {
	if f.askStreamFn == nil {
		return nil, errors.New("askStreamFn not set")
	}
	return f.askStreamFn(ctx, in, opts...)
}

type fakeConversationClient struct {
	listConversationsFn   func(ctx context.Context, in *pb.ListConversationsRequest, opts ...grpclib.CallOption) (*pb.ListConversationsResponse, error)
	getConversationFn     func(ctx context.Context, in *pb.GetConversationRequest, opts ...grpclib.CallOption) (*pb.GetConversationResponse, error)
	archiveConversationFn func(ctx context.Context, in *pb.ArchiveConversationRequest, opts ...grpclib.CallOption) (*pb.ArchiveConversationResponse, error)
}

func (f *fakeConversationClient) ListConversations(ctx context.Context, in *pb.ListConversationsRequest, opts ...grpclib.CallOption) (*pb.ListConversationsResponse, error) {
	if f.listConversationsFn == nil {
		return nil, errors.New("listConversationsFn not set")
	}
	return f.listConversationsFn(ctx, in, opts...)
}

func (f *fakeConversationClient) GetConversation(ctx context.Context, in *pb.GetConversationRequest, opts ...grpclib.CallOption) (*pb.GetConversationResponse, error) {
	if f.getConversationFn == nil {
		return nil, errors.New("getConversationFn not set")
	}
	return f.getConversationFn(ctx, in, opts...)
}

func (f *fakeConversationClient) ArchiveConversation(ctx context.Context, in *pb.ArchiveConversationRequest, opts ...grpclib.CallOption) (*pb.ArchiveConversationResponse, error) {
	if f.archiveConversationFn == nil {
		return nil, errors.New("archiveConversationFn not set")
	}
	return f.archiveConversationFn(ctx, in, opts...)
}

func TestInferenceGRPCClientAskMapsResponseAndAuthHeader(t *testing.T) {
	requestID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	sessionID := uuid.New()
	documentID := uuid.New()
	chunkID := uuid.New()

	c := &InferenceGRPCClient{
		timeout:   2 * time.Second,
		authToken: "test-token",
		client: &fakeInferenceClient{askFn: func(ctx context.Context, in *pb_inference.AskRequest, _ ...grpclib.CallOption) (*pb_inference.AskResponse, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}
			authValues := md.Get("authorization")
			if len(authValues) != 1 || authValues[0] != "Bearer test-token" {
				t.Fatalf("unexpected authorization header: %#v", authValues)
			}

			if in.GetRequestId() != requestID.String() {
				t.Fatalf("unexpected request id: %s", in.GetRequestId())
			}
			if in.GetUserId() != userID.String() {
				t.Fatalf("unexpected user id: %s", in.GetUserId())
			}
			if in.GetAccountId() != accountID.String() {
				t.Fatalf("unexpected account id: %s", in.GetAccountId())
			}

			title := "doc title"
			return &pb_inference.AskResponse{
				RequestId: requestID.String(),
				SessionId: sessionID.String(),
				Answer:    "Hello there",
				Model:     "gemini-1.5-flash",
				LatencyMs: 123,
				Citations: []*pb_inference.Citation{{
					DocumentId: documentID.String(),
					ChunkId:    chunkID.String(),
					SourceType: "chunk",
					Title:      &title,
					Score:      0.91,
				}},
				Usage: &pb_inference.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			}, nil
		}},
	}

	res, err := c.Ask(context.Background(), port.AskRequest{
		RequestID: requestID,
		UserID:    userID,
		AccountID: accountID,
		Query:     "hello",
		TopK:      5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.RequestID != requestID {
		t.Fatalf("unexpected response request id: %s", res.RequestID)
	}
	if res.SessionID != sessionID {
		t.Fatalf("unexpected response session id: %s", res.SessionID)
	}
	if res.Answer != "Hello there" {
		t.Fatalf("unexpected answer: %s", res.Answer)
	}
	if len(res.Citations) != 1 {
		t.Fatalf("unexpected citations count: %d", len(res.Citations))
	}
	if res.Citations[0].DocumentID != documentID || res.Citations[0].ChunkID != chunkID {
		t.Fatalf("unexpected citation ids: %+v", res.Citations[0])
	}
	if res.Usage.TotalTokens != 30 {
		t.Fatalf("unexpected usage: %+v", res.Usage)
	}
}

func TestInferenceGRPCClientAskRejectsOutOfRangeTopK(t *testing.T) {
	c := &InferenceGRPCClient{
		timeout: 1 * time.Second,
		client: &fakeInferenceClient{askFn: func(_ context.Context, _ *pb_inference.AskRequest, _ ...grpclib.CallOption) (*pb_inference.AskResponse, error) {
			t.Fatal("client call should not happen for invalid top_k")
			return nil, nil
		}},
	}

	_, err := c.Ask(context.Background(), port.AskRequest{
		RequestID: uuid.New(),
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "hi",
		TopK:      100,
	})
	if err == nil {
		t.Fatal("expected top_k validation error")
	}
}

func TestMapAskErrorMapsGRPCStatus(t *testing.T) {
	quotaErr := status.Error(codes.ResourceExhausted, "quota")
	err := mapAskError(quotaErr)
	if err == nil || err.Error() == "" {
		t.Fatal("expected mapped error")
	}
	if got := err.Error(); got == "quota" || got == "" {
		t.Fatalf("expected wrapped error message, got: %s", got)
	}
}

func TestInferenceGRPCClientAskStreamReturnsChunks(t *testing.T) {
	requestID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()

	streamClient := &fakeServerStream{chunks: []*pb_inference.AskStreamChunk{
		{Chunk: &pb_inference.AskStreamChunk_Text{Text: &pb_inference.TextChunk{Text: "Hello"}}},
		{Chunk: &pb_inference.AskStreamChunk_Text{Text: &pb_inference.TextChunk{Text: " World"}}},
		{Chunk: &pb_inference.AskStreamChunk_Citations{Citations: &pb_inference.CitationsChunk{Citations: []*pb_inference.Citation{
			{DocumentId: uuid.New().String(), ChunkId: uuid.New().String(), SourceType: "chunk", Score: 0.9},
		}}}},
		{Chunk: &pb_inference.AskStreamChunk_Done{Done: &pb_inference.DoneChunk{Model: "gemini-1.5-flash", LatencyMs: 100}}},
	}}

	c := &InferenceGRPCClient{
		timeout:   2 * time.Second,
		authToken: "",
		client: &fakeInferenceClient{askStreamFn: func(_ context.Context, _ *pb_inference.AskRequest, _ ...grpclib.CallOption) (grpclib.ServerStreamingClient[pb_inference.AskStreamChunk], error) {
			return streamClient, nil
		}},
	}

	ch, err := c.AskStream(context.Background(), port.AskRequest{
		RequestID: requestID,
		UserID:    userID,
		AccountID: accountID,
		Query:     "hi",
		TopK:      5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	chunks := make([]port.AskStreamChunk, 0)
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}
	if chunks[0].Text == nil || *chunks[0].Text != "Hello" {
		t.Fatalf("unexpected first chunk: %+v", chunks[0])
	}
	if chunks[1].Text == nil || *chunks[1].Text != " World" {
		t.Fatalf("unexpected second chunk: %+v", chunks[1])
	}
	if len(chunks[2].Citations) != 1 {
		t.Fatalf("unexpected citations chunk: %+v", chunks[2])
	}
	if chunks[3].Done == nil || chunks[3].Done.Model != "gemini-1.5-flash" {
		t.Fatalf("unexpected done chunk: %+v", chunks[3])
	}
}

func TestInferenceGRPCClientListConversations(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	sessionID := uuid.New()

	c := &InferenceGRPCClient{
		timeout:   2 * time.Second,
		authToken: "",
		conversationClient: &fakeConversationClient{listConversationsFn: func(_ context.Context, in *pb.ListConversationsRequest, _ ...grpclib.CallOption) (*pb.ListConversationsResponse, error) {
			if in.GetUserId() != userID.String() {
				t.Fatalf("unexpected user id: %s", in.GetUserId())
			}
			return &pb.ListConversationsResponse{
				Sessions: []*pb.SessionSummary{{
					Id:        sessionID.String(),
					AccountId: accountID.String(),
					Title:     "Test Chat",
					Language:  "en",
				}},
				Total: 1,
			}, nil
		}},
	}

	res, err := c.ListConversations(context.Background(), port.ListConversationsRequest{
		UserID:    userID,
		AccountID: accountID,
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Sessions) != 1 {
		t.Fatalf("unexpected sessions count: %d", len(res.Sessions))
	}
	if res.Sessions[0].Title != "Test Chat" {
		t.Fatalf("unexpected title: %s", res.Sessions[0].Title)
	}
	if res.Total != 1 {
		t.Fatalf("unexpected total: %d", res.Total)
	}
}

func TestInferenceGRPCClientGetConversation(t *testing.T) {
	sessionID := uuid.New()
	accountID := uuid.New()
	messageID := uuid.New()

	c := &InferenceGRPCClient{
		timeout:   2 * time.Second,
		authToken: "",
		conversationClient: &fakeConversationClient{getConversationFn: func(_ context.Context, in *pb.GetConversationRequest, _ ...grpclib.CallOption) (*pb.GetConversationResponse, error) {
			if in.GetSessionId() != sessionID.String() {
				t.Fatalf("unexpected session id: %s", in.GetSessionId())
			}
			return &pb.GetConversationResponse{
				Session: &pb.SessionDetail{
					Id:        sessionID.String(),
					AccountId: accountID.String(),
					Title:     "My Chat",
					Language:  "en",
				},
				Messages: []*pb.MessageDetail{{
					Id:        messageID.String(),
					Role:      "user",
					Content:   "Hello",
					CreatedAt: "2026-01-01T00:00:00Z",
				}},
				TotalMessages: 1,
			}, nil
		}},
	}

	res, err := c.GetConversation(context.Background(), port.GetConversationRequest{
		SessionID:      sessionID,
		AccountID:      accountID,
		MessageLimit:   50,
		MessageOffset:  0,
		IncludeDeleted: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Session.Title != "My Chat" {
		t.Fatalf("unexpected title: %s", res.Session.Title)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("unexpected messages count: %d", len(res.Messages))
	}
	if res.Messages[0].Content != "Hello" {
		t.Fatalf("unexpected message content: %s", res.Messages[0].Content)
	}
	if res.TotalMsgs != 1 {
		t.Fatalf("unexpected total: %d", res.TotalMsgs)
	}
}

func TestInferenceGRPCClientArchiveConversation(t *testing.T) {
	sessionID := uuid.New()
	accountID := uuid.New()

	c := &InferenceGRPCClient{
		timeout:   2 * time.Second,
		authToken: "",
		conversationClient: &fakeConversationClient{archiveConversationFn: func(_ context.Context, in *pb.ArchiveConversationRequest, _ ...grpclib.CallOption) (*pb.ArchiveConversationResponse, error) {
			if in.GetSessionId() != sessionID.String() {
				t.Fatalf("unexpected session id: %s", in.GetSessionId())
			}
			return &pb.ArchiveConversationResponse{}, nil
		}},
	}

	res, err := c.ArchiveConversation(context.Background(), port.ArchiveConversationRequest{
		SessionID: sessionID,
		AccountID: accountID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatal("expected success")
	}
}

type fakeServerStream struct {
	chunks []*pb_inference.AskStreamChunk
	index  int
}

func (f *fakeServerStream) Recv() (*pb_inference.AskStreamChunk, error) {
	if f.index >= len(f.chunks) {
		return nil, errors.New("EOF")
	}
	chunk := f.chunks[f.index]
	f.index++
	return chunk, nil
}

func (f *fakeServerStream) Header() (metadata.MD, error) { return nil, nil }
func (f *fakeServerStream) Trailer() metadata.MD         { return nil }
func (f *fakeServerStream) CloseSend() error             { return nil }
func (f *fakeServerStream) Context() context.Context     { return context.Background() }
func (f *fakeServerStream) SendMsg(m interface{}) error  { return nil }
func (f *fakeServerStream) RecvMsg(m interface{}) error  { return nil }
