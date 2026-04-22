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
	lastAskRequest  *port.AskRequest
}

func (m *mockInferencePort) Ask(ctx context.Context, req port.AskRequest) (port.AskResponse, error) {
	m.lastAskRequest = &req
	return m.askResult, m.askErr
}

func (m *mockInferencePort) AskStream(ctx context.Context, req port.AskRequest) (<-chan port.AskStreamChunk, error) {
	m.lastAskRequest = &req
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

func (m *mockInferencePort) ListConversations(context.Context, port.ListConversationsRequest) (port.ListConversationsResponse, error) {
	return port.ListConversationsResponse{}, nil
}

func (m *mockInferencePort) GetConversation(context.Context, port.GetConversationRequest) (port.GetConversationResponse, error) {
	return port.GetConversationResponse{}, nil
}

func (m *mockInferencePort) ArchiveConversation(context.Context, port.ArchiveConversationRequest) (port.ArchiveConversationResponse, error) {
	return port.ArchiveConversationResponse{Success: true, UpdatedAt: "2026-01-01T12:00:00Z"}, nil
}

func TestAskService_Ask_PassesTitleToPort(t *testing.T) {
	mock := &mockInferencePort{
		askResult: port.AskResponse{RequestID: uuid.New(), SessionID: uuid.New(), Answer: "ok"},
	}
	svc := NewAskService(mock)
	title := "Custom Title"

	_, err := svc.Ask(context.Background(), AskInput{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Query:     "test query",
		Language:  "en",
		TopK:      5,
		Title:     &title,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastAskRequest == nil || mock.lastAskRequest.Title == nil || *mock.lastAskRequest.Title != title {
		t.Fatal("expected title to be forwarded to inference port")
	}
}

func TestAskService_AskStream_ReturnsChunks(t *testing.T) {
	mock := &mockInferencePort{askStreamChunks: []port.AskStreamChunk{{Text: ptr("Hello")}, {Done: &port.DoneInfo{Model: "m"}}}}
	svc := NewAskService(mock)

	stream, err := svc.AskStream(context.Background(), AskStreamInput{UserID: uuid.New(), AccountID: uuid.New(), Query: "q"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := 0
	for range stream {
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 chunks, got %d", count)
	}
}

func TestAskService_ArchiveConversation(t *testing.T) {
	svc := NewAskService(&mockInferencePort{})
	out, err := svc.ArchiveConversation(context.Background(), ArchiveConversationInput{SessionID: uuid.New(), AccountID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Success {
		t.Fatal("expected success")
	}
	if out.UpdatedAt != "2026-01-01T12:00:00Z" {
		t.Fatalf("unexpected updatedAt: %s", out.UpdatedAt)
	}
}

func ptr[T any](v T) *T { return &v }
