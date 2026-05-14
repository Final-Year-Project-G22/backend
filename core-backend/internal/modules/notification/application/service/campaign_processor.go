package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CampaignProcessor struct {
	campaignTemplateRepo repository.CampaignTemplateRepository
	templateRend         *TemplateRenderer
	queueRepo            repository.NotificationQueueRepository
	accountReader        repository.AccountReader
	transactor           sharedrepo.Transactor
	logger               core.Logger
}

func NewCampaignProcessor(
	campaignTemplateRepo repository.CampaignTemplateRepository,
	templateRend *TemplateRenderer,
	queueRepo repository.NotificationQueueRepository,
	accountReader repository.AccountReader,
	transactor sharedrepo.Transactor,
	logger core.Logger,
) *CampaignProcessor {
	return &CampaignProcessor{
		campaignTemplateRepo: campaignTemplateRepo,
		templateRend:         templateRend,
		queueRepo:            queueRepo,
		accountReader:        accountReader,
		transactor:           transactor,
		logger:               logger,
	}
}

func (p *CampaignProcessor) ProcessCampaign(ctx context.Context, campaign *entity.NotificationCampaign) error {
	tmpl, err := p.campaignTemplateRepo.GetByID(ctx, campaign.CampaignTemplateID)
	if err != nil {
		return fmt.Errorf("failed to load campaign template %s: %w", campaign.CampaignTemplateID, err)
	}

	accountIDs, err := p.resolveRecipients(ctx, campaign)
	if err != nil {
		return fmt.Errorf("failed to resolve campaign recipients: %w", err)
	}

	channels := p.resolveChannels(tmpl)

	for _, accountID := range accountIDs {
		if err := p.enqueueForRecipient(ctx, campaign, tmpl, accountID, channels); err != nil {
			p.logger.Error("Failed to enqueue campaign notification for recipient",
				core.String("campaignID", campaign.ID.String()),
				core.String("accountID", accountID.String()),
				core.Error(err),
			)
			continue
		}
	}

	return nil
}

func (p *CampaignProcessor) ResolveSegment(ctx context.Context, campaignType entity.CampaignType, segment map[string]interface{}) ([]uuid.UUID, error) {
	switch campaignType {
	case entity.CampaignTypeBroadcast:
		return p.accountReader.FindAll(ctx)
	case entity.CampaignTypeSegmented:
		if len(segment) == 0 {
			return nil, fmt.Errorf("segment filters required for segmented campaign")
		}
		return p.accountReader.FindBySegment(ctx, segment)
	default:
		return nil, fmt.Errorf("unknown campaign type: %s", campaignType)
	}
}

func (p *CampaignProcessor) resolveRecipients(ctx context.Context, campaign *entity.NotificationCampaign) ([]uuid.UUID, error) {
	// If the campaign already has resolved account IDs (stored during scheduling), use those directly.
	if campaign.TargetSegment != nil {
		segment := *campaign.TargetSegment
		if resolvedRaw, ok := segment["resolvedAccountIDs"]; ok {
			if resolvedList, ok := resolvedRaw.([]interface{}); ok {
				accountIDs := make([]uuid.UUID, 0, len(resolvedList))
				for _, r := range resolvedList {
					if idStr, ok := r.(string); ok {
						id, err := uuid.Parse(idStr)
						if err == nil {
							accountIDs = append(accountIDs, id)
						}
					}
				}
				return accountIDs, nil
			}
			return nil, fmt.Errorf("invalid resolvedAccountIDs format")
		}
	}

	accountIDs, err := p.ResolveSegment(ctx, campaign.CampaignType, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve campaign segment: %w", err)
	}

	if len(accountIDs) == 0 {
		p.logger.Warn("Campaign resolved to zero recipients",
			core.String("campaignID", campaign.ID.String()),
		)
	}

	return accountIDs, nil
}

func (p *CampaignProcessor) resolveChannels(tmpl *entity.CampaignTemplate) []entity.Channel {
	content := tmpl.DefaultContent
	var channels []entity.Channel
	if content == nil {
		return channels
	}
	for key := range content {
		ch := entity.Channel(key)
		switch ch {
		case entity.ChannelEmail, entity.ChannelInApp, entity.ChannelPush, entity.ChannelSMS:
			channels = append(channels, ch)
		}
	}
	if !tmpl.EnablePushMirror {
		return channels
	}
	expanded := make([]entity.Channel, 0, len(channels)+1)
	for _, ch := range channels {
		expanded = append(expanded, ch)
		if ch == entity.ChannelInApp {
			expanded = append(expanded, entity.ChannelPush)
		}
	}
	return expanded
}

func isChannelContentValid(ch entity.Channel, content map[string]interface{}) bool {
	switch ch {
	case entity.ChannelEmail:
		subject, _ := content["subject"].(string)
		body, _ := content["body"].(string)
		return strings.TrimSpace(subject) != "" && strings.TrimSpace(body) != ""
	case entity.ChannelInApp:
		title, _ := content["title"].(string)
		body, _ := content["body"].(string)
		return strings.TrimSpace(title) != "" && strings.TrimSpace(body) != ""
	case entity.ChannelPush:
		title, _ := content["title"].(string)
		body, _ := content["body"].(string)
		return strings.TrimSpace(title) != "" && strings.TrimSpace(body) != ""
	}
	return false
}

func (p *CampaignProcessor) enqueueForRecipient(
	ctx context.Context,
	campaign *entity.NotificationCampaign,
	tmpl *entity.CampaignTemplate,
	accountID uuid.UUID,
	channels []entity.Channel,
) error {
	// 1. Resolve recipient account info (email + locale)
	accountInfo, err := p.accountReader.GetAccountInfo(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account info for %s: %w", accountID, err)
	}

	// 2. Resolve content by locale (try translation, fallback to default)
	resolvedContent := tmpl.DefaultContent
	if accountInfo.Locale != "" {
		translation, err := p.campaignTemplateRepo.GetTranslation(ctx, tmpl.ID, accountInfo.Locale)
		if err == nil {
			resolvedContent = translation.Content
		}
	}

	// 3. Render (no variable substitution for campaigns)
	rendered, err := p.templateRend.Render(resolvedContent, nil)
	if err != nil {
		return fmt.Errorf("failed to render content for account %s: %w", accountID, err)
	}

	// 4. For each channel, create queue entry
	for _, ch := range channels {
		channelContent, hasChannel := rendered[string(ch)]

		var channelMap map[string]interface{}

		switch {
		case !hasChannel && ch == entity.ChannelPush:
			inAppContent, hasInApp := rendered[string(entity.ChannelInApp)]
			if !hasInApp {
				continue
			}
			inAppMap, ok := inAppContent.(map[string]interface{})
			if !ok {
				continue
			}
			channelMap = map[string]interface{}{
				"title": inAppMap["title"],
				"body":  inAppMap["body"],
			}
			if actionUrl, ok := inAppMap["actionUrl"].(string); ok && actionUrl != "" {
				channelMap["actionUrl"] = actionUrl
			}
		case hasChannel:
			var ok bool
			channelMap, ok = channelContent.(map[string]interface{})
			if !ok {
				p.logger.Warn("Channel content is not a map, skipping",
					core.String("channel", string(ch)),
					core.String("accountID", accountID.String()),
				)
				continue
			}
		default:
			continue
		}

		if !isChannelContentValid(ch, channelMap) {
			p.logger.Warn("Skipping channel with empty required fields",
				core.String("channel", string(ch)),
				core.String("accountID", accountID.String()),
			)
			continue
		}

		var payload datatypes.JSONMap
		switch ch {
		case entity.ChannelEmail:
			payload = datatypes.JSONMap{
				"to":      accountInfo.Email,
				"subject": channelMap["subject"],
				"body":    channelMap["body"],
			}
		case entity.ChannelInApp:
			payload = datatypes.JSONMap{
				"title":   channelMap["title"],
				"content": channelMap["body"],
			}
			if actionUrl, ok := channelMap["actionUrl"].(string); ok && actionUrl != "" {
				payload["actionUrl"] = actionUrl
			}
		case entity.ChannelPush:
			payload = datatypes.JSONMap{
				"title": channelMap["title"],
				"body":  channelMap["body"],
			}
		}

		now := time.Now().UTC()
		queueItem := &entity.NotificationQueue{
			NotificationType: entity.NotificationTypeCampaign,
			AccountID:        accountID,
			Priority:         entity.NotificationPriorityMedium,
			Channel:          ch,
			Payload:          payload,
			ScheduledFor:     now,
			MaxRetries:       3,
			Status:           entity.NotificationStatusPending,
		}

		if err := p.queueRepo.Create(ctx, queueItem); err != nil {
			return fmt.Errorf("failed to create queue entry for account %s: %w", accountID, err)
		}
	}

	return nil
}
