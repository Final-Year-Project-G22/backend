package email

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/email"
	"gopkg.in/gomail.v2"
)

var _ repository.EmailProvider = (*SMTPProvider)(nil)

type SMTPProvider struct {
	config email.Config
	logger core.Logger
}

func NewSMTPProvider(config email.Config, logger core.Logger) *SMTPProvider {
	return &SMTPProvider{
		config: config,
		logger: logger,
	}
}

func (p *SMTPProvider) Send(ctx context.Context, to, subject, body string, metadata map[string]string) (string, error) {
	if !p.config.Enabled {
		p.logger.Warn("SMTP is disabled, skipping email send", core.String("to", to))
		return "", fmt.Errorf("smtp email provider is disabled")
	}

	msg := gomail.NewMessage()

	if p.config.FromName != "" {
		msg.SetHeader("From", fmt.Sprintf("%s <%s>", p.config.FromName, p.config.From))
	} else {
		msg.SetHeader("From", p.config.From)
	}

	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	// Add any metadata as headers
	for key, value := range metadata {
		msg.SetHeader(key, value)
	}

	dialer := gomail.NewDialer(
		p.config.Host,
		p.config.Port,
		p.config.Username,
		p.config.Password,
	)

	if p.config.EnableTLS {
		dialer.TLSConfig = &tls.Config{
			ServerName: p.config.Host,
			MinVersion: tls.VersionTLS12,
		}
	}

	if err := dialer.DialAndSend(msg); err != nil {
		p.logger.Error("SMTP send failed", core.String("to", to), core.Error(err))
		return "", fmt.Errorf("smtp send failed: %w", err)
	}

	// SMTP doesn't provide a message ID like Resend does
	return "", nil
}

func (p *SMTPProvider) VerifyWebhookSignature(_ context.Context, _ []byte, _, _ string) (bool, error) {
	// SMTP doesn't use webhooks, so always return true
	return true, nil
}
