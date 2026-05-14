package push

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"google.golang.org/api/option"
)

type FCMProvider struct {
	client *messaging.Client
	logger core.Logger
}

func NewFCMProvider(cfg core.FCMConfig, logger core.Logger) (*FCMProvider, error) {
	opt := option.WithCredentialsFile(cfg.CredentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get firebase messaging client: %w", err)
	}

	logger.Info("FCM provider initialized",
		core.String("credentialsFile", cfg.CredentialsFile),
	)

	return &FCMProvider{client: client, logger: logger}, nil
}

func (p *FCMProvider) Send(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	msg := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := p.client.Send(ctx, msg)
	if err != nil {
		p.logger.Error("FCM send failed",
			core.String("token", deviceToken),
			core.Error(err),
		)
		return fmt.Errorf("fcm send failed: %w", err)
	}

	return nil
}
