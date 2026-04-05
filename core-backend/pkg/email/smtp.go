package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

type smtpEmailer struct {
	dialer *gomail.Dialer
	config Config
}

func NewEmailer(cfg Config) (Emailer, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("email is not enabled")
	}

	dialer := gomail.NewDialer(
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
	)

	if cfg.EnableTLS {
		dialer.TLSConfig = &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
	}

	return &smtpEmailer{
		dialer: dialer,
		config: cfg,
	}, nil
}

func (s *smtpEmailer) Send(ctx context.Context, input SendInput) error {
	msg := gomail.NewMessage()

	from := input.From
	if from == "" {
		from = s.config.From
	}

	fromName := input.FromName
	if fromName == "" {
		fromName = s.config.FromName
	}

	if fromName != "" {
		msg.SetHeader("From", fmt.Sprintf("%s <%s>", fromName, from))
	} else {
		msg.SetHeader("From", from)
	}

	msg.SetHeader("To", input.To...)

	if len(input.CC) > 0 {
		msg.SetHeader("Cc", input.CC...)
	}

	if len(input.BCC) > 0 {
		msg.SetHeader("Bcc", input.BCC...)
	}

	msg.SetHeader("Subject", input.Subject)

	if input.IsHTML {
		msg.SetBody("text/html", input.Body)
	} else {
		msg.SetBody("text/plain", input.Body)
	}

	if err := s.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *smtpEmailer) SendTemplate(ctx context.Context, input SendTemplateInput) error {
	msg := gomail.NewMessage()

	from := input.From
	if from == "" {
		from = s.config.From
	}

	fromName := input.FromName
	if fromName == "" {
		fromName = s.config.FromName
	}

	if fromName != "" {
		msg.SetHeader("From", fmt.Sprintf("%s <%s>", fromName, from))
	} else {
		msg.SetHeader("From", from)
	}

	msg.SetHeader("To", input.To...)

	if len(input.CC) > 0 {
		msg.SetHeader("Cc", input.CC...)
	}

	if len(input.BCC) > 0 {
		msg.SetHeader("Bcc", input.BCC...)
	}

	msg.SetHeader("Subject", input.Subject)

	body := input.Template
	if input.Data != nil {
		body = s.replaceTemplateVariables(body, input.Data)
	}

	msg.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *smtpEmailer) replaceTemplateVariables(template string, data interface{}) string {
	result := template

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			placeholder := fmt.Sprintf("{{.%s}}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	case map[string]string:
		for key, value := range v {
			placeholder := fmt.Sprintf("{{.%s}}", key)
			result = strings.ReplaceAll(result, placeholder, value)
		}
	}

	return result
}
