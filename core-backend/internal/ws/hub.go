package ws

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	mu          sync.RWMutex
	clients     map[uuid.UUID]*Client
	subscribers map[uuid.UUID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[uuid.UUID]*Client),
		subscribers: make(map[uuid.UUID]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client, accountID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.clients[accountID]; ok {
		existing.Close()
	}
	h.clients[accountID] = client
}

func (h *Hub) Unregister(accountID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[accountID]; ok {
		for _, subs := range h.subscribers {
			delete(subs, client)
		}
		client.Close()
		delete(h.clients, accountID)
	}
}

func (h *Hub) Subscribe(client *Client, threadID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs, ok := h.subscribers[threadID]
	if !ok {
		subs = make(map[*Client]struct{})
		h.subscribers[threadID] = subs
	}
	subs[client] = struct{}{}
}

func (h *Hub) Unsubscribe(client *Client, threadID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.subscribers[threadID]; ok {
		delete(subs, client)
		if len(subs) == 0 {
			delete(h.subscribers, threadID)
		}
	}
}

func (h *Hub) PublishToThread(threadID uuid.UUID, event *ServerEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if subs, ok := h.subscribers[threadID]; ok {
		data, err := event.Marshal()
		if err != nil {
			return
		}
		for client := range subs {
			client.Send(data)
		}
	}
}

func (h *Hub) PublishToAll(event *ServerEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, err := event.Marshal()
	if err != nil {
		return
	}
	for _, client := range h.clients {
		client.Send(data)
	}
}
