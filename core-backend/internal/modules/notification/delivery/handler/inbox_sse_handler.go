package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

// resetFlusher wraps a gin.ResponseWriter to reset the write deadline on each flush.
// This prevents http.Server.WriteTimeout from killing long-lived SSE connections.
type resetFlusher struct {
	gin.ResponseWriter
}

func (w *resetFlusher) Flush() {
	ctrl := http.NewResponseController(w.ResponseWriter)
	_ = ctrl.SetWriteDeadline(time.Now().Add(30 * time.Second))
	w.ResponseWriter.Flush()
}

func (h *InboxSSEHandler) HandleInboxEvents(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if c.Request.Method == "OPTIONS" {
		c.Status(200)
		return
	}

	accountID, err := h.authenticate(c)
	if err != nil {
		h.logger.Error("Inbox SSE auth failed", core.Error(err))
		c.Status(401)
		return
	}

	h.logger.Info("Inbox SSE client connected", core.String("accountID", accountID.String()))

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Writer = &resetFlusher{ResponseWriter: c.Writer}

	// Flush the headers to establish the connection
	c.Writer.Flush()

	ch := h.broadcaster.Subscribe(accountID)
	defer func() {
		h.broadcaster.Unsubscribe(accountID, ch)
		h.logger.Info("Inbox SSE client disconnected", core.String("accountID", accountID.String()))
	}()

	flusher, ok := c.Writer.(interface{ Flush() })
	if !ok {
		h.logger.Error("ResponseWriter doesn't support flushing")
		c.Status(500)
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				h.logger.Error("Failed to marshal SSE event", core.Error(err))
				continue
			}
			_, _ = fmt.Fprintf(c.Writer, "event: notification_new\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(c.Writer, ": heartbeat\n\n")
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
