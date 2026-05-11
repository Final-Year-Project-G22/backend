package handler

import (
	"io"
	"net/http"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/usecase"
	"github.com/gin-gonic/gin"
)

// WebhookHandler handles Chapa webhook events.
type WebhookHandler struct {
	usecase usecase.PaymentUseCase
	logger  core.Logger
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(uc usecase.PaymentUseCase, logger core.Logger) *WebhookHandler {
	return &WebhookHandler{usecase: uc, logger: logger}
}

// HandleChapaWebhook receives and processes Chapa payment webhooks.
// This is a raw gin handler (not Huma) because we need raw body access for signature verification.
func (h *WebhookHandler) HandleChapaWebhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read webhook body", core.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	signature := c.GetHeader("x-chapa-signature")
	if signature == "" {
		signature = c.GetHeader("chapa-signature")
	}

	if err := h.usecase.HandleWebhook(c.Request.Context(), rawBody, signature); err != nil {
		h.logger.Error("webhook handling failed", core.Error(err))
		// Return 200 even on errors to prevent Chapa retries
		// unless it's a signature verification failure (401)
		if err.Error() == "invalid signature" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid signature"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
