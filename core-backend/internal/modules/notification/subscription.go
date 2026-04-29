package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/google/uuid"
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
						logger.Error("Failed to parse event",
							core.String("event", evt),
							core.Error(err),
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
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid event payload: %w", err)
	}

	accountID := parseUUID(raw["accountId"])
	nt := resolveNotificationType(eventType)

	if accountID == uuid.Nil || nt == "" {
		return nil, fmt.Errorf("unable to resolve account or notification type for event: %s", eventType)
	}

	variables := make(map[string]string)
	for k, v := range raw {
		if val, ok := v.(string); ok {
			variables[k] = val
		}
	}

	return &usecase.ProcessEventInput{
		SourceModule:     resolveSourceModule(eventType),
		SourceEvent:      eventType,
		NotificationType: nt,
		AccountID:        accountID,
		Variables:        variables,
	}, nil
}

func resolveNotificationType(eventType string) entity.NotificationType {
	switch eventType {
	case event.AccountRegistered, event.WelcomeMessage:
		return entity.NotificationTypeWelcomeMessage
	case event.AccountVerification:
		return entity.NotificationTypeAccountVerification
	case event.PasswordReset:
		return entity.NotificationTypePasswordReset
	case event.AccountAlert:
		return entity.NotificationTypeAccountAlert
	case event.ThreadReply:
		return entity.NotificationTypeCommunityReply
	case event.ThreadSolution:
		return entity.NotificationTypeCommunitySolution
	case event.ThreadMention:
		return entity.NotificationTypeCommunityMention
	case event.GuideStepCompleted:
		return entity.NotificationTypeGuideStepCompleted
	case event.GuideDeadline:
		return entity.NotificationTypeGuideDeadline
	case event.GuideUpdate:
		return entity.NotificationTypeGuideUpdate
	case event.AIQuotaLimit:
		return entity.NotificationTypeAIQuotaLimit
	case event.AIResponseReady:
		return entity.NotificationTypeAIResponseReady
	case event.PaymentConfirmation:
		return entity.NotificationTypePaymentConfirmation
	default:
		return ""
	}
}

func resolveSourceModule(eventType string) string {
	switch eventType {
	case event.AccountRegistered, event.AccountVerification, event.PasswordReset, event.AccountAlert, event.WelcomeMessage:
		return "iam"
	case event.ThreadReply, event.ThreadSolution, event.ThreadMention:
		return "community"
	case event.GuideStepCompleted, event.GuideDeadline, event.GuideUpdate:
		return "guide"
	case event.AIQuotaLimit, event.AIResponseReady:
		return "ai"
	case event.PaymentConfirmation:
		return "payment"
	default:
		return "unknown"
	}
}

func parseUUID(v interface{}) uuid.UUID {
	switch val := v.(type) {
	case string:
		id, err := uuid.Parse(val)
		if err == nil {
			return id
		}
	case map[string]interface{}:
		if s, ok := val["String"].(string); ok {
			id, err := uuid.Parse(s)
			if err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
