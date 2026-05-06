package handler

import (
	"encoding/json"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	broadcaster *service.CampaignSSEBroadcaster
}

func NewSSEHandler(broadcaster *service.CampaignSSEBroadcaster) *SSEHandler {
	return &SSEHandler{broadcaster: broadcaster}
}

func (h *SSEHandler) HandleCampaignEvents(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(ch)

	flusher, ok := c.Writer.(interface{ Flush() })
	if !ok {
		c.Status(500)
		return
	}

	for {
		select {
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: campaign_status\ndata: %s\n\n", data)
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
