package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type notificationIngestUsecase struct {
	tmplRepo      repository.NotificationTemplateRepository
	prefRepo      repository.UserNotificationPreferenceRepository
	mutedRepo     repository.MutedAccountRepository
	queueRepo     repository.NotificationQueueRepository
	muteResolvers []repository.MuteResolver
	renderer      *service.TemplateRenderer
	transactor    sharedrepo.Transactor
}

func NewNotificationIngestUsecase(
	tmplRepo repository.NotificationTemplateRepository,
	prefRepo repository.UserNotificationPreferenceRepository,
	mutedRepo repository.MutedAccountRepository,
	queueRepo repository.NotificationQueueRepository,
	renderer *service.TemplateRenderer,
	transactor sharedrepo.Transactor,
	muteResolvers ...repository.MuteResolver,
) usecase.NotificationIngestUsecase {
	return &notificationIngestUsecase{
		tmplRepo:      tmplRepo,
		prefRepo:      prefRepo,
		mutedRepo:     mutedRepo,
		queueRepo:     queueRepo,
		renderer:      renderer,
		transactor:    transactor,
		muteResolvers: muteResolvers,
	}
}

func (uc *notificationIngestUsecase) ProcessEvent(ctx context.Context, input usecase.ProcessEventInput) error {
	tmpl, err := uc.tmplRepo.GetByType(ctx, input.NotificationType)
	if err != nil {
		return err
	}

	if err := uc.validateVariables(tmpl, input.Variables); err != nil {
		return err
	}

	channels := uc.channelsFromTemplate(tmpl)
	if len(channels) == 0 {
		return nil
	}

	mutedAccountID := uc.extractMutedAccountID(input.Metadata)

	for _, channel := range channels {
		if !uc.hasChannelContent(tmpl, channel) {
			continue
		}

		allowed, err := uc.isChannelAllowed(ctx, input.AccountID, input.NotificationType, channel)
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}

		if mutedAccountID != nil {
			muted, err := uc.mutedRepo.IsMuted(ctx, input.AccountID, *mutedAccountID)
			if err != nil {
				return err
			}
			if muted {
				continue
			}
		}

		muted, err := uc.isMutedByResolvers(ctx, input.AccountID, input.Metadata)
		if err != nil {
			return err
		}
		if muted {
			continue
		}

		if err := uc.enqueue(ctx, tmpl, channel, input.AccountID, input.Variables, nil); err != nil {
			return err
		}
	}

	return nil
}

func (uc *notificationIngestUsecase) SendNotification(ctx context.Context, input usecase.SendNotificationInput) error {
	tmpl, err := uc.tmplRepo.GetByType(ctx, input.NotificationType)
	if err != nil {
		return err
	}

	if err := uc.validateVariables(tmpl, input.Variables); err != nil {
		return err
	}

	if !uc.hasChannelContent(tmpl, input.Channel) {
		return notiferror.ErrInvalidChannel
	}

	allowed, err := uc.isChannelAllowed(ctx, input.AccountID, input.NotificationType, input.Channel)
	if err != nil {
		return err
	}
	if !allowed {
		return nil
	}

	mutedAccountID := uc.extractMutedAccountID(input.Metadata)
	if mutedAccountID != nil {
		muted, err := uc.mutedRepo.IsMuted(ctx, input.AccountID, *mutedAccountID)
		if err != nil {
			return err
		}
		if muted {
			return nil
		}
	}

	muted, err := uc.isMutedByResolvers(ctx, input.AccountID, input.Metadata)
	if err != nil {
		return err
	}
	if muted {
		return nil
	}

	return uc.enqueue(ctx, tmpl, input.Channel, input.AccountID, input.Variables, input.ScheduledFor)
}

func (uc *notificationIngestUsecase) SendMultiChannel(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, variables map[string]string, metadata map[string]interface{}, channels []entity.Channel, expiresAt *time.Time) error {
	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		tmpl, err := uc.tmplRepo.GetByType(txCtx, notificationType)
		if err != nil {
			return err
		}

		if err := uc.validateVariables(tmpl, variables); err != nil {
			return err
		}

		mutedAccountID := uc.extractMutedAccountID(metadata)

		for _, channel := range channels {
			if !uc.hasChannelContent(tmpl, channel) {
				continue
			}

			allowed, err := uc.isChannelAllowed(txCtx, accountID, notificationType, channel)
			if err != nil {
				return err
			}
			if !allowed {
				continue
			}

			if mutedAccountID != nil {
				muted, err := uc.mutedRepo.IsMuted(txCtx, accountID, *mutedAccountID)
				if err != nil {
					return err
				}
				if muted {
					continue
				}
			}

			muted, err := uc.isMutedByResolvers(txCtx, accountID, metadata)
			if err != nil {
				return err
			}
			if muted {
				continue
			}

			if err := uc.enqueue(txCtx, tmpl, channel, accountID, variables, nil); err != nil {
				return err
			}
		}

		return nil
	})
}

func (uc *notificationIngestUsecase) enqueue(
	ctx context.Context,
	tmpl *entity.NotificationTemplate,
	channel entity.Channel,
	accountID uuid.UUID,
	variables map[string]string,
	scheduledFor *time.Time,
) error {
	contentMap := map[string]interface{}(tmpl.DefaultContent)
	rendered, err := uc.renderer.RenderMultiChannel(contentMap, variables, []entity.Channel{channel})
	if err != nil {
		return fmt.Errorf("failed to render content for channel %s: %w", channel, err)
	}

	payload, ok := rendered[string(channel)]
	if !ok {
		return fmt.Errorf("no rendered content for channel %s", channel)
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected payload type for channel %s", channel)
	}

	item := &entity.NotificationQueue{
		NotificationType: tmpl.NotificationType,
		AccountID:        accountID,
		Priority:         tmpl.Priority,
		TemplateID:       &tmpl.ID,
		Channel:          channel,
		Payload:          datatypes.JSONMap(payloadMap),
		ScheduledFor:     uc.scheduledTime(scheduledFor),
		MaxRetries:       3,
		RetryCount:       0,
		Status:           entity.NotificationStatusPending,
	}

	return uc.queueRepo.Create(ctx, item)
}

func (uc *notificationIngestUsecase) validateVariables(tmpl *entity.NotificationTemplate, variables map[string]string) error {
	if tmpl.VariablesSchema == nil || variables == nil {
		return nil
	}
	schema := map[string]interface{}(*tmpl.VariablesSchema)
	return uc.renderer.ValidateVariables(schema, variables)
}

func (uc *notificationIngestUsecase) isChannelAllowed(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (bool, error) {
	pref, err := uc.prefRepo.GetByAccountAndTypeAndChannel(ctx, accountID, notificationType, channel)
	if err != nil {
		if err == notiferror.ErrPreferenceNotFound {
			return true, nil
		}
		return false, err
	}
	if !pref.IsEnabled {
		return false, nil
	}
	if pref.QuietHoursStart != nil && pref.QuietHoursEnd != nil {
		now := time.Now()
		nowTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
		start := time.Date(0, 1, 1, pref.QuietHoursStart.Hour(), pref.QuietHoursStart.Minute(), pref.QuietHoursStart.Second(), 0, pref.QuietHoursStart.Location())
		end := time.Date(0, 1, 1, pref.QuietHoursEnd.Hour(), pref.QuietHoursEnd.Minute(), pref.QuietHoursEnd.Second(), 0, pref.QuietHoursEnd.Location())
		if nowTime.After(start) && nowTime.Before(end) {
			return false, nil
		}
	}
	return true, nil
}

func (uc *notificationIngestUsecase) isMutedByResolvers(ctx context.Context, accountID uuid.UUID, metadata map[string]interface{}) (bool, error) {
	if len(uc.muteResolvers) == 0 || metadata == nil {
		return false, nil
	}
	itemType, _ := metadata["itemType"].(string)
	itemIDStr, _ := metadata["itemId"].(string)
	if itemType == "" || itemIDStr == "" {
		return false, nil
	}
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return false, nil
	}
	for _, resolver := range uc.muteResolvers {
		muted, err := resolver.IsMuted(ctx, accountID, itemType, itemID)
		if err != nil {
			return false, err
		}
		if muted {
			return true, nil
		}
	}
	return false, nil
}

func (uc *notificationIngestUsecase) extractMutedAccountID(metadata map[string]interface{}) *uuid.UUID {
	if metadata == nil {
		return nil
	}
	idStr, ok := metadata["mutedAccountId"].(string)
	if !ok {
		return nil
	}
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return nil
	}
	return &parsed
}

func (uc *notificationIngestUsecase) channelsFromTemplate(tmpl *entity.NotificationTemplate) []entity.Channel {
	var channels []entity.Channel
	for key := range tmpl.DefaultContent {
		channels = append(channels, entity.Channel(key))
	}
	return channels
}

func (uc *notificationIngestUsecase) hasChannelContent(tmpl *entity.NotificationTemplate, channel entity.Channel) bool {
	_, ok := tmpl.DefaultContent[string(channel)]
	return ok
}

func (uc *notificationIngestUsecase) scheduledTime(scheduledFor *time.Time) time.Time {
	if scheduledFor != nil {
		return *scheduledFor
	}
	return time.Now().UTC()
}
