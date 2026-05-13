package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	iamservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InboxSSEHandler struct {
	broadcaster  *service.InboxSSEBroadcaster
	tokenService token.TokenService
	authService  iamservice.AuthService
	logger       core.Logger
}

func NewInboxSSEHandler(broadcaster *service.InboxSSEBroadcaster, tokenService token.TokenService, authService iamservice.AuthService, logger core.Logger) *InboxSSEHandler {
	return &InboxSSEHandler{
		broadcaster:  broadcaster,
		tokenService: tokenService,
		authService:  authService,
		logger:       logger,
	}
}

func (h *InboxSSEHandler) HandleInboxEvents(c *gin.Context) {
	accountID, err := h.authenticate(c)
	if err != nil {
		c.Status(401)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := h.broadcaster.Subscribe(accountID)
	defer h.broadcaster.Unsubscribe(accountID, ch)

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
			_, _ = fmt.Fprintf(c.Writer, "event: inbox_new\ndata: %s\n\n", data)
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func (h *InboxSSEHandler) authenticate(c *gin.Context) (uuid.UUID, error) {
	authHeader := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		tokenStr = c.Query("token")
		if tokenStr == "" {
			return uuid.Nil, fmt.Errorf("missing authorization")
		}
	}

	claims, err := h.tokenService.ValidateAccessToken(c.Request.Context(), tokenStr)
	if err != nil {
		h.logger.Error("Inbox SSE auth failed", core.Error(err))
		return uuid.Nil, err
	}

	accountID, err := h.authService.GetAccountIDBySessionID(c.Request.Context(), claims.SessionID)
	if err != nil {
		h.logger.Error("Failed to get account ID from session", core.Error(err))
		return uuid.Nil, err
	}

	return accountID, nil
}
