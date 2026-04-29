package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/event"
	"github.com/Final-Year-Project-G22/backend/core/pkg/email"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

type EventHandler func(ctx context.Context, data []byte) error

func RegisterSubscribers(bus interface {
	Subscribe(event string, handler func(context.Context, []byte) error) error
}, emailer email.Emailer, logger Logger) error {
	if bus == nil || emailer == nil {
		logger.Info("Skipping event handler registration - bus or emailer not available")
		return nil
	}

	handlers := map[string]EventHandler{
		event.AccountRegistered: func(ctx context.Context, data []byte) error {
			return handleUserRegistered(ctx, data, emailer)
		},
		event.UserEmailOTPRequested: func(ctx context.Context, data []byte) error {
			return handleUserEmailOTPRequested(ctx, data, emailer)
		},
		event.AdminCreated: func(ctx context.Context, data []byte) error {
			return handleAdminCreated(ctx, data, emailer)
		},
	}

	for eventName, handler := range handlers {
		err := bus.Subscribe(eventName, handler)
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventName, err)
		}
		logger.Info("Subscribed to event", zap.String("event", eventName))
	}

	return nil
}

func handleUserRegistered(ctx context.Context, data []byte, emailer email.Emailer) error {
	var evt event.AccountRegisteredEvent

	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal user registered event: %w", err)
	}

	fmt.Println("EVENT RECEIVED: user.registered for", evt.Email)

	locale := evt.Locale
	if locale == "" || !i18n.HasLocale(locale) {
		locale = "en"
	}

	subject := i18n.Resolve("email.welcome.subject", locale)
	templateBody := i18n.Resolve("email.welcome.template", locale)

	err := emailer.SendTemplate(ctx, email.SendTemplateInput{
		To:       []string{evt.Email},
		Subject:  subject,
		Template: templateBody,
		Data: map[string]string{
			"firstName": evt.FirstName,
		},
	})

	if err != nil {
		fmt.Println("EMAIL FAILED:", err)
		return err
	}

	fmt.Println("EMAIL SENT to", evt.Email)

	return nil
}

func handleUserEmailOTPRequested(ctx context.Context, data []byte, emailer email.Emailer) error {
	var evt event.UserEmailOTPRequestedEvent

	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal user email otp requested event: %w", err)
	}

	locale := evt.Locale
	if locale == "" || !i18n.HasLocale(locale) {
		locale = "en"
	}

	subject := i18n.Resolve("email.verification.subject", locale)
	templateBody := i18n.Resolve("email.verification.template", locale)

	err := emailer.SendTemplate(ctx, email.SendTemplateInput{
		To:       []string{evt.Email},
		Subject:  subject,
		Template: templateBody,
		Data: map[string]string{
			"firstName":      evt.FirstName,
			"otpCode":        evt.OTPCode,
			"expiresMinutes": fmt.Sprintf("%d", evt.ExpiresMinutes),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func handleAdminCreated(ctx context.Context, data []byte, emailer email.Emailer) error {
	var evt event.AdminCreatedEvent

	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal admin credentials event: %w", err)
	}

	locale := evt.Locale
	if locale == "" || !i18n.HasLocale(locale) {
		locale = "en"
	}

	subject := i18n.Resolve("email.adminCredentials.subject", locale)
	templateBody := i18n.Resolve("email.adminCredentials.template", locale)

	return emailer.SendTemplate(ctx, email.SendTemplateInput{
		To:       []string{evt.Email},
		Subject:  subject,
		Template: templateBody,
		Data: map[string]string{
			"firstName": evt.FirstName,
			"lastName":  evt.LastName,
			"email":     evt.Email,
			"password":  evt.Password,
		},
	})
}
