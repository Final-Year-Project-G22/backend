package service

import (
	"context"
	"sync"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type InboxEvent struct {
	Type             string                  `json:"type"`
	AccountID        uuid.UUID               `json:"accountId"`
	NotificationType entity.NotificationType `json:"notificationType"`
	Timestamp        time.Time               `json:"timestamp"`
}

type inboxSSEClient struct {
	accountID uuid.UUID
	ch        chan InboxEvent
}

type InboxSSEBroadcaster struct {
	clients    map[uuid.UUID]map[chan InboxEvent]struct{}
	register   chan *inboxSSEClient
	unregister chan *inboxSSEClient
	broadcast  chan InboxEvent
	mu         sync.Mutex
	logger     core.Logger
}

func NewInboxSSEBroadcaster(logger core.Logger) *InboxSSEBroadcaster {
	return &InboxSSEBroadcaster{
		clients:    make(map[uuid.UUID]map[chan InboxEvent]struct{}),
		register:   make(chan *inboxSSEClient),
		unregister: make(chan *inboxSSEClient),
		broadcast:  make(chan InboxEvent, 100),
		logger:     logger,
	}
}

func (b *InboxSSEBroadcaster) Start(ctx context.Context) {
	go b.run(ctx)
	b.logger.Info("Inbox SSE broadcaster started")
}

func (b *InboxSSEBroadcaster) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-b.register:
			b.mu.Lock()
			if _, ok := b.clients[client.accountID]; !ok {
				b.clients[client.accountID] = make(map[chan InboxEvent]struct{})
			}
			b.clients[client.accountID][client.ch] = struct{}{}
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			if clients, ok := b.clients[client.accountID]; ok {
				delete(clients, client.ch)
				close(client.ch)
				if len(clients) == 0 {
					delete(b.clients, client.accountID)
				}
			}
			b.mu.Unlock()
		case event := <-b.broadcast:
			b.mu.Lock()
			if clients, ok := b.clients[event.AccountID]; ok {
				for ch := range clients {
					select {
					case ch <- event:
					default:
						b.logger.Warn("Inbox SSE client channel full, dropping event",
							core.String("accountID", event.AccountID.String()),
						)
					}
				}
			}
			b.mu.Unlock()
		}
	}
}

func (b *InboxSSEBroadcaster) Subscribe(accountID uuid.UUID) chan InboxEvent {
	ch := make(chan InboxEvent, 10)
	b.register <- &inboxSSEClient{accountID: accountID, ch: ch}
	return ch
}

func (b *InboxSSEBroadcaster) Unsubscribe(accountID uuid.UUID, ch chan InboxEvent) {
	b.unregister <- &inboxSSEClient{accountID: accountID, ch: ch}
}

func (b *InboxSSEBroadcaster) Publish(event InboxEvent) {
	b.broadcast <- event
}
