package email

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
)

type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string, metadata map[string]string) (providerMessageID string, err error)
	VerifyWebhookSignature(ctx context.Context, payload []byte, svixTimestamp, svixSignature string) (bool, error)
}

type ResendProvider struct {
	config     core.ResendConfig
	httpClient *http.Client
	logger     core.Logger
}

func NewResendProvider(config core.ResendConfig, logger core.Logger) *ResendProvider {
	return &ResendProvider{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type resendSendRequest struct {
	From    string            `json:"from"`
	To      []string          `json:"to"`
	Subject string            `json:"subject"`
	HTML    string            `json:"html"`
	Headers map[string]string `json:"headers,omitempty"`
}

type resendSendResponse struct {
	ID string `json:"id"`
}

func (p *ResendProvider) Send(ctx context.Context, to, subject, body string, metadata map[string]string) (string, error) {
	if !p.config.Enabled {
		p.logger.Warn("Resend is disabled, skipping email send", core.String("to", to))
		return "", fmt.Errorf("resend email provider is disabled")
	}

	reqBody := resendSendRequest{
		From:    fmt.Sprintf("%s <%s>", p.config.FromName, p.config.FromEmail),
		To:      []string{to},
		Subject: subject,
		HTML:    body,
		Headers: metadata,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal resend request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Error("Resend API request failed", core.String("to", to), core.Error(err))
		return "", fmt.Errorf("resend API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read resend response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Resend API returned non-200",
			core.Int("status", resp.StatusCode),
			core.String("body", string(respBody)),
		)
		return "", fmt.Errorf("resend API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var sendResp resendSendResponse
	if err := json.Unmarshal(respBody, &sendResp); err != nil {
		return "", fmt.Errorf("failed to parse resend response: %w", err)
	}

	return sendResp.ID, nil
}

func (p *ResendProvider) VerifyWebhookSignature(_ context.Context, payload []byte, svixTimestamp, svixSignature string) (bool, error) {
	if p.config.WebhookSecret == "" {
		return false, fmt.Errorf("resend webhook secret is not configured")
	}

	secretBytes, err := base64.StdEncoding.DecodeString(p.config.WebhookSecret)
	if err != nil {
		p.logger.Error("Failed to decode resend webhook secret", core.Error(err))
		return false, fmt.Errorf("invalid resend webhook secret: %w", err)
	}

	timestamp, err := strconv.ParseInt(svixTimestamp, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid svix timestamp: %w", err)
	}

	now := time.Now().Unix()
	if now-timestamp > 300 {
		p.logger.Warn("Resend webhook timestamp is too old")
		return false, nil
	}

	signedContent := fmt.Sprintf("%s.%s", svixTimestamp, string(payload))

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(signedContent))
	expectedMAC := mac.Sum(nil)

	signatures := strings.Split(svixSignature, " ")
	for _, sig := range signatures {
		sig = strings.TrimSpace(sig)
		if sig == "" {
			continue
		}
		parts := strings.SplitN(sig, "=", 2)
		if len(parts) != 2 {
			continue
		}
		version := parts[0]
		sigValue := parts[1]

		decodedSig, err := base64.StdEncoding.DecodeString(sigValue)
		if err != nil {
			decodedSig, err = hex.DecodeString(sigValue)
			if err != nil {
				continue
			}
		}

		if hmac.Equal(decodedSig, expectedMAC) && version == "v1" {
			return true, nil
		}
	}

	return false, nil
}
