package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	emailUC   usecase.EmailDeliveryUsecase
	emailProv repository.EmailProvider
	logger    core.Logger
}

func NewWebhookHandler(
	emailUC usecase.EmailDeliveryUsecase,
	emailProv repository.EmailProvider,
	logger core.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		emailUC:   emailUC,
		emailProv: emailProv,
		logger:    logger,
	}
}

func (h *WebhookHandler) HandleResendWebhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook request body", core.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	svixSignature := c.GetHeader("svix-signature-256")
	svixTimestamp := c.GetHeader("svix-timestamp")

	if svixSignature == "" || svixTimestamp == "" {
		h.logger.Warn("Webhook missing svix headers")
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error"})
		return
	}

	valid, err := h.emailProv.VerifyWebhookSignature(c.Request.Context(), rawBody, svixTimestamp, svixSignature)
	if err != nil || !valid {
		h.logger.Warn("Webhook signature verification failed", core.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error"})
		return
	}

	payload, err := parseResendWebhook(rawBody)
	if err != nil {
		h.logger.Error("Failed to parse webhook payload", core.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	eventType := mapResendEventType(payload.Type)
	if eventType == "" {
		h.logger.Warn("Unknown resend webhook event type", core.String("type", payload.Type))
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, payload.Data.CreatedAt)
	if err != nil {
		occurredAt = time.Now().UTC()
	}

	event := usecase.ResendWebhookEvent{
		EventType:      eventType,
		EmailID:        payload.Data.EmailID,
		RecipientEmail: getFirstRecipient(payload.Data.To),
		OccurredAt:     occurredAt,
		BounceReason:   payload.Data.BounceReason,
	}

	if err := h.emailUC.HandleWebhookEvent(c.Request.Context(), event); err != nil {
		h.logger.Error("Failed to handle webhook event",
			core.String("eventType", eventType),
			core.String("emailID", payload.Data.EmailID),
			core.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type rawResendWebhook struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func parseResendWebhook(raw []byte) (*dto.ResendWebhookPayload, error) {
	var rawPayload rawResendWebhook
	if err := json.Unmarshal(raw, &rawPayload); err != nil {
		return nil, err
	}

	emailID, _ := rawPayload.Data["email_id"].(string)
	toRaw, _ := rawPayload.Data["to"].([]interface{})
	var to []string
	for _, t := range toRaw {
		if s, ok := t.(string); ok {
			to = append(to, s)
		}
	}
	subject, _ := rawPayload.Data["subject"].(string)
	createdAt, _ := rawPayload.Data["created_at"].(string)
	bounceReasonRaw, _ := rawPayload.Data["bounce_reason"].(string)
	var bounceReason *string
	if bounceReasonRaw != "" {
		bounceReason = &bounceReasonRaw
	}

	return &dto.ResendWebhookPayload{
		Type: rawPayload.Type,
		Data: dto.ResendWebhookData{
			EmailID:      emailID,
			To:           to,
			Subject:      subject,
			CreatedAt:    createdAt,
			BounceReason: bounceReason,
		},
	}, nil
}

func mapResendEventType(resendType string) string {
	switch resendType {
	case "email.delivered":
		return "delivered"
	case "email.opened":
		return "opened"
	case "email.clicked":
		return "clicked"
	case "email.bounced":
		return "bounced"
	case "email.complained":
		return "complained"
	default:
		return ""
	}
}

func getFirstRecipient(to []string) string {
	if len(to) > 0 {
		return to[0]
	}
	return ""
}
