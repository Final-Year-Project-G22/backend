package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/google/uuid"
)

const defaultConversationCacheTTL = 5 * time.Minute

type CachingInferencePort struct {
	base  port.AIInferencePort
	cache port.ConversationCachePort
	ttl   time.Duration

	mu                sync.Mutex
	listKeysByAccount map[uuid.UUID]map[string]struct{}
	getKeysBySession  map[uuid.UUID]map[string]struct{}
}

func NewCachingInferencePort(base port.AIInferencePort, cache port.ConversationCachePort, ttl time.Duration) *CachingInferencePort {
	if ttl <= 0 {
		ttl = defaultConversationCacheTTL
	}
	return &CachingInferencePort{
		base:              base,
		cache:             cache,
		ttl:               ttl,
		listKeysByAccount: make(map[uuid.UUID]map[string]struct{}),
		getKeysBySession:  make(map[uuid.UUID]map[string]struct{}),
	}
}

func (c *CachingInferencePort) Ask(ctx context.Context, req port.AskRequest) (port.AskResponse, error) {
	resp, err := c.base.Ask(ctx, req)
	if err == nil {
		c.invalidateListKeysByAccount(ctx, req.AccountID)
	}
	return resp, err
}

func (c *CachingInferencePort) AskStream(ctx context.Context, req port.AskRequest) (<-chan port.AskStreamChunk, error) {
	ch, err := c.base.AskStream(ctx, req)
	if err == nil {
		c.invalidateListKeysByAccount(ctx, req.AccountID)
	}
	return ch, err
}

func (c *CachingInferencePort) ListConversations(ctx context.Context, req port.ListConversationsRequest) (port.ListConversationsResponse, error) {
	key := listConversationsCacheKey(req)

	var cached port.ListConversationsResponse
	if hit, err := c.cache.Get(ctx, key, &cached); err == nil && hit {
		return cached, nil
	}

	resp, err := c.base.ListConversations(ctx, req)
	if err != nil {
		return port.ListConversationsResponse{}, err
	}

	_ = c.cache.Set(ctx, key, resp, c.ttl)
	c.registerListKey(req.AccountID, key)

	return resp, nil
}

func (c *CachingInferencePort) GetConversation(ctx context.Context, req port.GetConversationRequest) (port.GetConversationResponse, error) {
	key := getConversationCacheKey(req)

	var cached port.GetConversationResponse
	if hit, err := c.cache.Get(ctx, key, &cached); err == nil && hit {
		return cached, nil
	}

	resp, err := c.base.GetConversation(ctx, req)
	if err != nil {
		return port.GetConversationResponse{}, err
	}

	_ = c.cache.Set(ctx, key, resp, c.ttl)
	c.registerGetKey(req.SessionID, key)

	return resp, nil
}

func (c *CachingInferencePort) ArchiveConversation(ctx context.Context, req port.ArchiveConversationRequest) (port.ArchiveConversationResponse, error) {
	resp, err := c.base.ArchiveConversation(ctx, req)
	if err != nil {
		return port.ArchiveConversationResponse{}, err
	}

	c.invalidateGetKeysBySession(ctx, req.SessionID)
	c.invalidateListKeysByAccount(ctx, req.AccountID)

	return resp, nil
}

func (c *CachingInferencePort) registerListKey(accountID uuid.UUID, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.listKeysByAccount[accountID]; !ok {
		c.listKeysByAccount[accountID] = make(map[string]struct{})
	}
	c.listKeysByAccount[accountID][key] = struct{}{}
}

func (c *CachingInferencePort) registerGetKey(sessionID uuid.UUID, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.getKeysBySession[sessionID]; !ok {
		c.getKeysBySession[sessionID] = make(map[string]struct{})
	}
	c.getKeysBySession[sessionID][key] = struct{}{}
}

func (c *CachingInferencePort) invalidateListKeysByAccount(ctx context.Context, accountID uuid.UUID) {
	keys := c.popListKeys(accountID)
	if len(keys) == 0 {
		return
	}
	_ = c.cache.Invalidate(ctx, keys...)
}

func (c *CachingInferencePort) invalidateGetKeysBySession(ctx context.Context, sessionID uuid.UUID) {
	keys := c.popGetKeys(sessionID)
	if len(keys) == 0 {
		return
	}
	_ = c.cache.Invalidate(ctx, keys...)
}

func (c *CachingInferencePort) popListKeys(accountID uuid.UUID) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	set := c.listKeysByAccount[accountID]
	delete(c.listKeysByAccount, accountID)
	return keySetToSlice(set)
}

func (c *CachingInferencePort) popGetKeys(sessionID uuid.UUID) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	set := c.getKeysBySession[sessionID]
	delete(c.getKeysBySession, sessionID)
	return keySetToSlice(set)
}

func keySetToSlice(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}

func listConversationsCacheKey(req port.ListConversationsRequest) string {
	return fmt.Sprintf("ai:conv:list:%s:%s:%d:%d", req.AccountID, req.UserID, req.Limit, req.Offset)
}

func getConversationCacheKey(req port.GetConversationRequest) string {
	return fmt.Sprintf("ai:conv:get:%s:%s:%d:%d:%t", req.AccountID, req.SessionID, req.MessageLimit, req.MessageOffset, req.IncludeDeleted)
}
