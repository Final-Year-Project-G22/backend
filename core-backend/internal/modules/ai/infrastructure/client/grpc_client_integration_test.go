package client

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	pb "github.com/Final-Year-Project-G22/backend/core/pb/ai/inference/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const testBufSize = 1024 * 1024

type testInferenceServer struct {
	pb.UnimplementedAIInferenceServiceServer
}

func (s *testInferenceServer) Ask(_ context.Context, req *pb.AskRequest) (*pb.AskResponse, error) {
	if req.GetQuery() == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	requestID := req.GetRequestId()
	sessionID := uuid.New().String()
	documentID := uuid.New().String()
	chunkID := uuid.New().String()
	title := "business license"

	return &pb.AskResponse{
		RequestId: requestID,
		SessionId: sessionID,
		Answer:    "You should start with licensing.",
		Model:     "mock-model",
		LatencyMs: 25,
		Citations: []*pb.Citation{{
			DocumentId: documentID,
			ChunkId:    chunkID,
			SourceType: "chunk",
			Title:      &title,
			Score:      0.88,
		}},
		Usage: &pb.Usage{PromptTokens: 8, CompletionTokens: 13, TotalTokens: 21},
	}, nil
}

func TestInferenceGRPCClientAskOverTransport(t *testing.T) {
	lis := bufconn.Listen(testBufSize)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		authValues := md.Get("authorization")
		if len(authValues) != 1 || authValues[0] != "Bearer valid-token" {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}))
	t.Cleanup(grpcServer.Stop)

	pb.RegisterAIInferenceServiceServer(grpcServer, &testInferenceServer{})

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("create grpc conn: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := &InferenceGRPCClient{
		client:    pb.NewAIInferenceServiceClient(conn),
		timeout:   2 * time.Second,
		authToken: "valid-token",
	}

	requestID := uuid.New()
	res, err := client.Ask(context.Background(), port.AskRequest{
		RequestID: requestID,
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "How can I open a shop?",
		Language:  constants.LocaleEnglish,
		TopK:      5,
	})
	if err != nil {
		t.Fatalf("ask failed: %v", err)
	}

	if res.RequestID != requestID {
		t.Fatalf("unexpected request id: %s", res.RequestID)
	}
	if res.Answer == "" {
		t.Fatal("expected answer")
	}
	if len(res.Citations) != 1 {
		t.Fatalf("expected one citation, got %d", len(res.Citations))
	}
	if res.Usage.TotalTokens != 21 {
		t.Fatalf("unexpected usage total tokens: %d", res.Usage.TotalTokens)
	}
}

func TestInferenceGRPCClientAskMapsUnauthenticated(t *testing.T) {
	lis := bufconn.Listen(testBufSize)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("authorization")) == 0 || md.Get("authorization")[0] != "Bearer valid-token" {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}))
	t.Cleanup(grpcServer.Stop)

	pb.RegisterAIInferenceServiceServer(grpcServer, &testInferenceServer{})

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("create grpc conn: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := &InferenceGRPCClient{
		client:    pb.NewAIInferenceServiceClient(conn),
		timeout:   2 * time.Second,
		authToken: "wrong-token",
	}

	_, err = client.Ask(context.Background(), port.AskRequest{
		RequestID: uuid.New(),
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "hello",
		Language:  constants.LocaleEnglish,
		TopK:      5,
	})
	if err == nil {
		t.Fatal("expected authentication error")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
