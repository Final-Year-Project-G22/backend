package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/google/uuid"
)

type fakeInferencePort struct {
	listCalls    int
	getCalls     int
	askCalls     int
	archiveCalls int

	listResp    port.ListConversationsResponse
	getResp     port.GetConversationResponse
	archiveResp port.ArchiveConversationResponse
}

func (f *fakeInferencePort) Ask(context.Context, port.AskRequest) (port.AskResponse, error) {
	f.askCalls++
	return port.AskResponse{}, nil
}

func (f *fakeInferencePort) AskStream(context.Context, port.AskRequest) (<-chan port.AskStreamChunk, error) {
	ch := make(chan port.AskStreamChunk)
	close(ch)
	return ch, nil
}

func (f *fakeInferencePort) ListConversations(context.Context, port.ListConversationsRequest) (port.ListConversationsResponse, error) {
	f.listCalls++
	return f.listResp, nil
}

func (f *fakeInferencePort) GetConversation(context.Context, port.GetConversationRequest) (port.GetConversationResponse, error) {
	f.getCalls++
	return f.getResp, nil
}

func (f *fakeInferencePort) ArchiveConversation(context.Context, port.ArchiveConversationRequest) (port.ArchiveConversationResponse, error) {
	f.archiveCalls++
	return f.archiveResp, nil
}

type fakeConversationCache struct {
	values        map[string]string
	invalidateOps int
}

func newFakeConversationCache() *fakeConversationCache {
	return &fakeConversationCache{values: make(map[string]string)}
}

func (f *fakeConversationCache) Get(_ context.Context, key string, out any) (bool, error) {
	v, ok := f.values[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal([]byte(v), out); err != nil {
		return false, err
	}
	return true, nil
}

func (f *fakeConversationCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.values[key] = string(b)
	return nil
}

func (f *fakeConversationCache) Invalidate(_ context.Context, keys ...string) error {
	f.invalidateOps++
	for _, key := range keys {
		delete(f.values, key)
	}
	return nil
}

func TestCachingInferencePort_ListConversationsCacheHit(t *testing.T) {
	base := &fakeInferencePort{listResp: port.ListConversationsResponse{Total: 1}}
	cache := newFakeConversationCache()
	cached := NewCachingInferencePort(base, cache, time.Minute)

	req := port.ListConversationsRequest{AccountID: uuid.New(), UserID: uuid.New(), Limit: 20, Offset: 0}
	_, _ = cached.ListConversations(context.Background(), req)
	_, _ = cached.ListConversations(context.Background(), req)

	if base.listCalls != 1 {
		t.Fatalf("expected list base call once, got %d", base.listCalls)
	}
}

func TestCachingInferencePort_GetConversationCacheHit(t *testing.T) {
	base := &fakeInferencePort{getResp: port.GetConversationResponse{TotalMsgs: 2}}
	cache := newFakeConversationCache()
	cached := NewCachingInferencePort(base, cache, time.Minute)

	req := port.GetConversationRequest{AccountID: uuid.New(), SessionID: uuid.New(), MessageLimit: 50, MessageOffset: 0, IncludeDeleted: false}
	_, _ = cached.GetConversation(context.Background(), req)
	_, _ = cached.GetConversation(context.Background(), req)

	if base.getCalls != 1 {
		t.Fatalf("expected get base call once, got %d", base.getCalls)
	}
}

func TestCachingInferencePort_ArchiveInvalidatesConversationCache(t *testing.T) {
	base := &fakeInferencePort{archiveResp: port.ArchiveConversationResponse{Success: true}}
	cache := newFakeConversationCache()
	cached := NewCachingInferencePort(base, cache, time.Minute)

	accountID := uuid.New()
	sessionID := uuid.New()
	getReq := port.GetConversationRequest{AccountID: accountID, SessionID: sessionID, MessageLimit: 50, MessageOffset: 0, IncludeDeleted: false}

	_, _ = cached.GetConversation(context.Background(), getReq)
	_, _ = cached.ArchiveConversation(context.Background(), port.ArchiveConversationRequest{AccountID: accountID, SessionID: sessionID})
	_, _ = cached.GetConversation(context.Background(), getReq)

	if base.getCalls != 2 {
		t.Fatalf("expected get base call twice after invalidation, got %d", base.getCalls)
	}
	if cache.invalidateOps == 0 {
		t.Fatal("expected cache invalidation to be called")
	}
}

func TestCachingInferencePort_AskInvalidatesListCache(t *testing.T) {
	base := &fakeInferencePort{listResp: port.ListConversationsResponse{Total: 1}}
	cache := newFakeConversationCache()
	cached := NewCachingInferencePort(base, cache, time.Minute)

	accountID := uuid.New()
	userID := uuid.New()
	listReq := port.ListConversationsRequest{AccountID: accountID, UserID: userID, Limit: 20, Offset: 0}

	_, _ = cached.ListConversations(context.Background(), listReq)
	_, _ = cached.Ask(context.Background(), port.AskRequest{AccountID: accountID, UserID: userID, RequestID: uuid.New()})
	_, _ = cached.ListConversations(context.Background(), listReq)

	if base.listCalls != 2 {
		t.Fatalf("expected list base call twice after ask invalidation, got %d", base.listCalls)
	}
}
