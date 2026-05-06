package service

import (
	"context"
	"sync"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type CampaignStatusEvent struct {
	CampaignID uuid.UUID             `json:"campaignId"`
	Status     entity.CampaignStatus `json:"status"`
	Timestamp  time.Time             `json:"timestamp"`
}

type CampaignSSEBroadcaster struct {
	clients    map[chan CampaignStatusEvent]struct{}
	register   chan chan CampaignStatusEvent
	unregister chan chan CampaignStatusEvent
	broadcast  chan CampaignStatusEvent
	mu         sync.Mutex
	logger     core.Logger
}

func NewCampaignSSEBroadcaster(logger core.Logger) *CampaignSSEBroadcaster {
	return &CampaignSSEBroadcaster{
		clients:    make(map[chan CampaignStatusEvent]struct{}),
		register:   make(chan chan CampaignStatusEvent),
		unregister: make(chan chan CampaignStatusEvent),
		broadcast:  make(chan CampaignStatusEvent, 100),
		logger:     logger,
	}
}

func (b *CampaignSSEBroadcaster) Start(ctx context.Context) {
	go b.run(ctx)
	b.logger.Info("Campaign SSE broadcaster started")
}

func (b *CampaignSSEBroadcaster) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = struct{}{}
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			delete(b.clients, client)
			close(client)
			b.mu.Unlock()
		case event := <-b.broadcast:
			b.mu.Lock()
			for client := range b.clients {
				select {
				case client <- event:
				default:
					b.logger.Warn("SSE client channel full, dropping event",
						core.String("campaignID", event.CampaignID.String()),
						core.String("status", string(event.Status)),
					)
				}
			}
			b.mu.Unlock()
		}
	}
}

func (b *CampaignSSEBroadcaster) Subscribe() chan CampaignStatusEvent {
	ch := make(chan CampaignStatusEvent, 10)
	b.register <- ch
	return ch
}

func (b *CampaignSSEBroadcaster) Unsubscribe(ch chan CampaignStatusEvent) {
	b.unregister <- ch
}

func (b *CampaignSSEBroadcaster) Publish(event CampaignStatusEvent) {
	b.broadcast <- event
}
