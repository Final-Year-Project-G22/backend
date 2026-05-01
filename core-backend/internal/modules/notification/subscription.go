package notification

import (
	"context"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"go.uber.org/fx"
)

func registerEventSubscriptions(lc fx.Lifecycle, bus rabbitmq.Bus, ingestUC usecase.NotificationIngestUsecase, logger core.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			events := []string{
				event.AccountRegistered,
				event.AccountVerification,
				event.PasswordReset,
				event.AccountAlert,
				event.WelcomeMessage,
				event.ThreadReply,
				event.ThreadSolution,
				event.ThreadMention,
				event.GuideStepCompleted,
				event.GuideDeadline,
				event.GuideUpdate,
				event.AIQuotaLimit,
				event.AIResponseReady,
				event.PaymentConfirmation,
			}
			for _, e := range events {
				evt := e
				if err := bus.Subscribe(evt, func(ctx context.Context, data []byte) error {
					input, err := parseProcessEventInput(evt, data)
					if err != nil {
						logger.Error("Canonical event rejected",
							core.String("event", evt),
							core.String("reason", err.Error()),
						)
						return nil
					}
					return ingestUC.ProcessEvent(ctx, *input)
				}); err != nil {
					return fmt.Errorf("failed to subscribe to %s: %w", evt, err)
				}
			}
			logger.Info("Subscribed to notification events", core.Int("count", len(events)))
			return nil
		},
	})
}

func parseProcessEventInput(eventType string, data []byte) (*usecase.ProcessEventInput, error) {
	envelope, err := notificationevent.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("canonical event rejected: %w", err)
	}

	var channel *entity.Channel
	if envelope.Channel != nil {
		c := entity.Channel(*envelope.Channel)
		channel = &c
	}

	return &usecase.ProcessEventInput{
		SourceModule:     envelope.SourceModule,
		SourceEvent:      envelope.EventType,
		NotificationType: entity.NotificationType(envelope.NotificationType),
		ChannelPolicy:    string(envelope.ChannelPolicy),
		Channel:          channel,
		AccountID:        envelope.AccountID,
		Variables:        envelope.Variables,
		Metadata:         envelope.MetadataMap(),
	}, nil
}
