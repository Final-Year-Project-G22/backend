package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	pb "github.com/Final-Year-Project-G22/backend/core/pb/ai/inference/v1"
	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeInferenceClient struct {
	askFn       func(ctx context.Context, in *pb.AskRequest, opts ...grpclib.CallOption) (*pb.AskResponse, error)
	askStreamFn func(ctx context.Context, in *pb.AskRequest, opts ...grpclib.CallOption) (grpclib.ServerStreamingClient[pb.AskStreamChunk], error)
}

func (f *fakeInferenceClient) Ask(
	ctx context.Context,
	in *pb.AskRequest,
	opts ...grpclib.CallOption,
) (*pb.AskResponse, error) {
	if f.askFn == nil {
		return nil, errors.New("askFn not set")
	}
	return f.askFn(ctx, in, opts...)
}

func (f *fakeInferenceClient) AskStream(
	ctx context.Context,
	in *pb.AskRequest,
	opts ...grpclib.CallOption,
) (grpclib.ServerStreamingClient[pb.AskStreamChunk], error) {
	if f.askStreamFn == nil {
		return nil, errors.New("askStreamFn not set")
	}
	return f.askStreamFn(ctx, in, opts...)
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
		client: &fakeInferenceClient{askFn: func(ctx context.Context, in *pb.AskRequest, _ ...grpclib.CallOption) (*pb.AskResponse, error) {
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
			return &pb.AskResponse{
				RequestId: requestID.String(),
				SessionId: sessionID.String(),
				Answer:    "Hello there",
				Model:     "gemini-1.5-flash",
				LatencyMs: 123,
				Citations: []*pb.Citation{{
					DocumentId: documentID.String(),
					ChunkId:    chunkID.String(),
					SourceType: "chunk",
					Title:      &title,
					Score:      0.91,
				}},
				Usage: &pb.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
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
		client: &fakeInferenceClient{askFn: func(_ context.Context, _ *pb.AskRequest, _ ...grpclib.CallOption) (*pb.AskResponse, error) {
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
