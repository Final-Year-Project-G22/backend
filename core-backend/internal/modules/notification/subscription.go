package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	appservice "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"go.uber.org/fx"
)

func registerEventSubscriptions(lc fx.Lifecycle, bus rabbitmq.Bus, ingestUC usecase.NotificationIngestUsecase, logger core.Logger, syncSvc *appservice.SyncComplianceService, complianceRepo notifrepo.ComplianceEntryRepository, bpRepo iamrepo.BusinessProfileRepository) {
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
				event.UserEmailOTPRequested,
				event.ComplianceAlert,
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

	// Subscribe to business.profile.updated for compliance sync
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return bus.Subscribe(event.BusinessProfileUpdated, func(ctx context.Context, data []byte) error {
				envelope, err := notificationevent.Parse(data)
				if err != nil {
					logger.Error("Failed to parse business.profile.updated event", core.Error(err))
					return nil
				}
				return syncSvc.Sync(ctx, envelope.AccountID)
			})
		},
	})

	// Subscribe to guide.compliance_step_completed — create compliance entry
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return bus.Subscribe(event.GuideComplianceStepCompleted, func(ctx context.Context, data []byte) error {
				envelope, err := notificationevent.Parse(data)
				if err != nil {
					logger.Error("Failed to parse guide.compliance_step_completed event", core.Error(err))
					return nil
				}
				return handleComplianceStepCompleted(ctx, envelope, complianceRepo, bpRepo, logger)
			})
		},
	})

	// Subscribe to guide.compliance_step_rolled_back — delete compliance entry
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return bus.Subscribe(event.GuideComplianceStepRolledBack, func(ctx context.Context, data []byte) error {
				envelope, err := notificationevent.Parse(data)
				if err != nil {
					logger.Error("Failed to parse guide.compliance_step_rolled_back event", core.Error(err))
					return nil
				}
				return handleComplianceStepRolledBack(ctx, envelope, complianceRepo, logger)
			})
		},
	})
}

func handleComplianceStepCompleted(ctx context.Context, envelope *notificationevent.Envelope, repo notifrepo.ComplianceEntryRepository, bpRepo iamrepo.BusinessProfileRepository, logger core.Logger) error {
	complianceType := envelope.Variables["compliance_type"]
	if complianceType == "" {
		return nil
	}

	ctype := entity.ComplianceType(complianceType)
	defaults, ok := service.ComplianceDefaults[ctype]
	if !ok {
		logger.Warn("Unknown compliance type", core.String("complianceType", complianceType))
		return nil
	}

	profile, err := bpRepo.GetByAccountID(ctx, envelope.AccountID)
	if err != nil || profile == nil {
		logger.Warn("No business profile found for account",
			core.String("accountID", envelope.AccountID.String()))
		return nil
	}

	expiry := time.Now().UTC().AddDate(0, 0, defaults.ExpiryDurationDays)
	entry := &entity.ComplianceEntry{
		BusinessProfileID:  profile.ID,
		AccountID:          envelope.AccountID,
		ComplianceType:     ctype,
		Source:             entity.ComplianceSourceAuto,
		ExpiryDate:         expiry,
		ReminderDaysBefore: defaults.ReminderDays,
		Status:             entity.ComplianceEntryStatusActive,
	}
	return repo.Create(ctx, entry)
}

func handleComplianceStepRolledBack(ctx context.Context, envelope *notificationevent.Envelope, repo notifrepo.ComplianceEntryRepository, logger core.Logger) error {
	complianceType := envelope.Variables["compliance_type"]
	if complianceType == "" {
		return nil
	}
	return repo.DeleteByAccountAndType(ctx, envelope.AccountID, entity.ComplianceType(complianceType), entity.ComplianceSourceAuto)
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
