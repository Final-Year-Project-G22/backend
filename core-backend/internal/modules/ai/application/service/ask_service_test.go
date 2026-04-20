package service

import (
	"context"
	"testing"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/google/uuid"
)

type mockInferencePort struct {
	askResult       port.AskResponse
	askErr          error
	askStreamChunks []port.AskStreamChunk
	askStreamErr    error
}

func (m *mockInferencePort) Ask(ctx context.Context, req port.AskRequest) (port.AskResponse, error) {
	return m.askResult, m.askErr
}

func (m *mockInferencePort) AskStream(ctx context.Context, req port.AskRequest) (<-chan port.AskStreamChunk, error) {
	if m.askStreamErr != nil {
		return nil, m.askStreamErr
	}
	ch := make(chan port.AskStreamChunk, len(m.askStreamChunks))
	for _, c := range m.askStreamChunks {
		ch <- c
	}
	close(ch)
	return ch, nil
}

func (m *mockInferencePort) ListConversations(ctx context.Context, req port.ListConversationsRequest) (port.ListConversationsResponse, error) {
	return port.ListConversationsResponse{}, nil
}

func (m *mockInferencePort) GetConversation(ctx context.Context, req port.GetConversationRequest) (port.GetConversationResponse, error) {
	return port.GetConversationResponse{}, nil
}

func (m *mockInferencePort) ArchiveConversation(ctx context.Context, req port.ArchiveConversationRequest) (port.ArchiveConversationResponse, error) {
	return port.ArchiveConversationResponse{}, nil
}

func TestAskService_Ask_ReturnsResponse(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	sessionID := uuid.New()

	mock := &mockInferencePort{
		askResult: port.AskResponse{
			RequestID: uuid.New(),
			SessionID: sessionID,
			Answer:    "Test answer",
			Citations: []port.Citation{
				{
					DocumentID: uuid.New(),
					ChunkID:    uuid.New(),
					SourceType: "chunk",
					Score:      0.95,
				},
			},
			Usage:     port.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			Model:     "gemini-1.5-flash",
			LatencyMS: 100,
		},
	}

	svc := NewAskService(mock)
	resp, err := svc.Ask(context.Background(), AskInput{
		UserID:    userID,
		AccountID: accountID,
		Query:     "test query",
		Language:  "en",
		TopK:      5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer != "Test answer" {
		t.Errorf("expected answer 'Test answer', got %q", resp.Answer)
	}

	if len(resp.Citations) != 1 {
		t.Errorf("expected 1 citation, got %d", len(resp.Citations))
	}

	if resp.Model != "gemini-1.5-flash" {
		t.Errorf("expected model 'gemini-1.5-flash', got %q", resp.Model)
	}
}

func TestAskService_Ask_DefaultLanguage(t *testing.T) {
	mock := &mockInferencePort{}
	svc := NewAskService(mock)

	_, err := svc.Ask(context.Background(), AskInput{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "test",
		Language:  "",
		TopK:      5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskService_Ask_DefaultTopK(t *testing.T) {
	mock := &mockInferencePort{}
	svc := NewAskService(mock)

	_, err := svc.Ask(context.Background(), AskInput{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "test",
		TopK:      0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskService_AskStream_ReturnsChunks(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	mock := &mockInferencePort{
		askStreamChunks: []port.AskStreamChunk{
			{Text: ptr("Hello")},
			{Text: ptr(" World")},
			{
				Citations: []port.Citation{
					{DocumentID: uuid.New(), ChunkID: uuid.New(), SourceType: "chunk", Score: 0.9},
				},
			},
			{
				Done: &port.DoneInfo{
					Model:     "gemini-1.5-flash",
					Usage:     port.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
					LatencyMs: 150,
				},
			},
		},
	}

	svc := NewAskService(mock)
	stream, err := svc.AskStream(context.Background(), AskStreamInput{
		UserID:    userID,
		AccountID: accountID,
		Query:     "test query",
		Language:  "en",
		TopK:      5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	chunks := make([]port.AskStreamChunk, 0)
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 4 {
		t.Errorf("expected 4 chunks, got %d", len(chunks))
	}

	if chunks[0].Text == nil || *chunks[0].Text != "Hello" {
		t.Errorf("first chunk should be 'Hello', got %v", chunks[0].Text)
	}

	if chunks[1].Text == nil || *chunks[1].Text != " World" {
		t.Errorf("second chunk should be ' World', got %v", chunks[1].Text)
	}

	if len(chunks[2].Citations) != 1 {
		t.Errorf("third chunk should have 1 citation, got %d", len(chunks[2].Citations))
	}

	if chunks[3].Done == nil {
		t.Error("fourth chunk should be done")
	}
}

func TestAskService_AskStream_Error(t *testing.T) {
	mock := &mockInferencePort{
		askStreamErr: context.DeadlineExceeded,
	}

	svc := NewAskService(mock)
	_, err := svc.AskStream(context.Background(), AskStreamInput{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "test",
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAskService_ListConversations(t *testing.T) {
	mock := &mockInferencePort{}
	svc := NewAskService(mock)

	resp, err := svc.ListConversations(context.Background(), ListConversationsInput{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Limit:     20,
		Offset:    0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("expected 0 total, got %d", resp.Total)
	}
}

func TestAskService_GetConversation(t *testing.T) {
	mock := &mockInferencePort{}
	svc := NewAskService(mock)

	resp, err := svc.GetConversation(context.Background(), GetConversationInput{
		SessionID:      uuid.New(),
		AccountID:      uuid.New(),
		MessageLimit:   50,
		MessageOffset:  0,
		IncludeDeleted: false,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalMsgs != 0 {
		t.Errorf("expected 0 total msgs, got %d", resp.TotalMsgs)
	}
}

func TestAskService_ArchiveConversation(t *testing.T) {
	mock := &mockInferencePort{}
	svc := NewAskService(mock)

	resp, err := svc.ArchiveConversation(context.Background(), ArchiveConversationInput{
		SessionID: uuid.New(),
		AccountID: uuid.New(),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Success {
		t.Error("expected success")
	}
}

func ptr[T any](v T) *T {
	return &v
}
