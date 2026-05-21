package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type notificationIngestUsecase struct {
	tmplRepo      notifrepo.NotificationTemplateRepository
	prefRepo      notifrepo.UserNotificationPreferenceRepository
	mutedRepo     notifrepo.MutedAccountRepository
	queueRepo     notifrepo.NotificationQueueRepository
	accountRepo   iamrepo.AccountRepository
	muteResolvers []notifrepo.MuteResolver
	renderer      *service.TemplateRenderer
	transactor    sharedrepo.Transactor
}

func NewNotificationIngestUsecase(
	tmplRepo notifrepo.NotificationTemplateRepository,
	prefRepo notifrepo.UserNotificationPreferenceRepository,
	mutedRepo notifrepo.MutedAccountRepository,
	queueRepo notifrepo.NotificationQueueRepository,
	accountRepo iamrepo.AccountRepository,
	renderer *service.TemplateRenderer,
	transactor sharedrepo.Transactor,
	muteResolvers ...notifrepo.MuteResolver,
) usecase.NotificationIngestUsecase {
	return &notificationIngestUsecase{
		tmplRepo:      tmplRepo,
		prefRepo:      prefRepo,
		mutedRepo:     mutedRepo,
		queueRepo:     queueRepo,
		accountRepo:   accountRepo,
		renderer:      renderer,
		transactor:    transactor,
		muteResolvers: muteResolvers,
	}
}

func (uc *notificationIngestUsecase) resolveContent(ctx context.Context, tmpl *entity.NotificationTemplate, locale string) map[string]interface{} {
	if locale == "" || locale == "en" {
		return map[string]interface{}(tmpl.DefaultContent)
	}
	translations, err := uc.tmplRepo.GetTranslations(ctx, tmpl.ID)
	if err != nil {
		return map[string]interface{}(tmpl.DefaultContent)
	}
	for _, t := range translations {
		if t.Language == locale {
			return map[string]interface{}(t.Content)
		}
	}
	return map[string]interface{}(tmpl.DefaultContent)
}

func (uc *notificationIngestUsecase) ProcessEvent(ctx context.Context, input usecase.ProcessEventInput) error {
	tmpl, err := uc.tmplRepo.GetByType(ctx, input.NotificationType)
	if err != nil {
		if errors.Is(err, notiferror.ErrTemplateNotFound) {
			return rabbitmq.NewPermanentError(err)
		}
		return err
	}

	if err := uc.validateVariables(tmpl, input.Variables); err != nil {
		return err
	}

	locale, _ := input.Metadata["locale"].(string)
	contentMap := uc.resolveContent(ctx, tmpl, locale)

	channels := uc.resolveChannels(contentMap, input.ChannelPolicy, input.Channel)
	if len(channels) == 0 {
		return nil
	}

	isMuted, err := uc.resolveMuted(ctx, input.AccountID, input.Metadata)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if !uc.hasChannelContent(contentMap, channel) {
			continue
		}

		allowed, err := uc.isChannelAllowed(ctx, input.AccountID, input.NotificationType, channel)
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}

		if err := uc.enqueue(ctx, tmpl, contentMap, channel, input.AccountID, input.Variables, nil, isMuted); err != nil {
			return err
		}
	}

	return nil
}

func (uc *notificationIngestUsecase) resolveChannels(content map[string]interface{}, channelPolicy string, singleChannel *entity.Channel) []entity.Channel {
	if channelPolicy == "single" && singleChannel != nil {
		if _, ok := content[string(*singleChannel)]; ok {
			return []entity.Channel{*singleChannel}
		}
		return nil
	}
	return uc.channelsFromContent(content)
}

func (uc *notificationIngestUsecase) SendNotification(ctx context.Context, input usecase.SendNotificationInput) error {
	tmpl, err := uc.tmplRepo.GetByType(ctx, input.NotificationType)
	if err != nil {
		return err
	}

	if err := uc.validateVariables(tmpl, input.Variables); err != nil {
		return err
	}

	contentMap := map[string]interface{}(tmpl.DefaultContent)

	if !uc.hasChannelContent(contentMap, input.Channel) {
		return notiferror.ErrInvalidChannel
	}

	allowed, err := uc.isChannelAllowed(ctx, input.AccountID, input.NotificationType, input.Channel)
	if err != nil {
		return err
	}
	if !allowed {
		return nil
	}

	isMuted, err := uc.resolveMuted(ctx, input.AccountID, input.Metadata)
	if err != nil {
		return err
	}

	return uc.enqueue(ctx, tmpl, contentMap, input.Channel, input.AccountID, input.Variables, input.ScheduledFor, isMuted)
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

		contentMap := map[string]interface{}(tmpl.DefaultContent)

		isMuted, err := uc.resolveMuted(txCtx, accountID, metadata)
		if err != nil {
			return err
		}

		for _, channel := range channels {
			if !uc.hasChannelContent(contentMap, channel) {
				continue
			}

			allowed, err := uc.isChannelAllowed(txCtx, accountID, notificationType, channel)
			if err != nil {
				return err
			}
			if !allowed {
				continue
			}

			if err := uc.enqueue(txCtx, tmpl, contentMap, channel, accountID, variables, nil, isMuted); err != nil {
				return err
			}
		}

		return nil
	})
}

func (uc *notificationIngestUsecase) enqueue(
	ctx context.Context,
	tmpl *entity.NotificationTemplate,
	content map[string]interface{},
	channel entity.Channel,
	accountID uuid.UUID,
	variables map[string]string,
	scheduledFor *time.Time,
	isMuted bool,
) error {
	rendered, err := uc.renderer.RenderMultiChannel(content, variables, []entity.Channel{channel})
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

	if isMuted {
		payloadMap["_isMuted"] = true
	}

	account, err := uc.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return rabbitmq.NewPermanentError(fmt.Errorf("failed to get account email: %w", err))
	}
	payloadMap["to"] = account.Email

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

func (uc *notificationIngestUsecase) resolveMuted(ctx context.Context, accountID uuid.UUID, metadata map[string]interface{}) (bool, error) {
	mutedAccountID := uc.extractMutedAccountID(metadata)
	if mutedAccountID != nil {
		muted, err := uc.mutedRepo.IsMuted(ctx, accountID, *mutedAccountID)
		if err != nil {
			return false, err
		}
		if muted {
			return true, nil
		}
	}
	return uc.isMutedByResolvers(ctx, accountID, metadata)
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

func (uc *notificationIngestUsecase) channelsFromContent(content map[string]interface{}) []entity.Channel {
	var channels []entity.Channel
	for key := range content {
		channels = append(channels, entity.Channel(key))
	}
	return channels
}

func (uc *notificationIngestUsecase) hasChannelContent(content map[string]interface{}, channel entity.Channel) bool {
	_, ok := content[string(channel)]
	return ok
}

func (uc *notificationIngestUsecase) scheduledTime(scheduledFor *time.Time) time.Time {
	if scheduledFor != nil {
		return *scheduledFor
	}
	return time.Now().UTC()
}
