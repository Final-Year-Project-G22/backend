package dto

type ResendWebhookData struct {
	EmailID      string   `json:"email_id"`
	To           []string `json:"to"`
	Subject      string   `json:"subject"`
	CreatedAt    string   `json:"created_at"`
	BounceReason *string  `json:"bounce_reason,omitempty"`
}

type ResendWebhookPayload struct {
	Type string            `json:"type"`
	Data ResendWebhookData `json:"data"`
}

type WebhookResponse struct {
	Status string `json:"status"`
}
