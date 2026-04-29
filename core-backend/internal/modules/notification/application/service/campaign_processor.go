package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CampaignProcessor struct {
	templateUC   usecase.NotificationTemplateUsecase
	templateRend *TemplateRenderer
	queueRepo    repository.NotificationQueueRepository
	accountRepo  repository.AccountReader
	transactor   sharedrepo.Transactor
	ingestUC     usecase.NotificationIngestUsecase
	logger       core.Logger
}

func NewCampaignProcessor(
	templateUC usecase.NotificationTemplateUsecase,
	templateRend *TemplateRenderer,
	queueRepo repository.NotificationQueueRepository,
	accountRepo repository.AccountReader,
	transactor sharedrepo.Transactor,
	ingestUC usecase.NotificationIngestUsecase,
	logger core.Logger,
) *CampaignProcessor {
	return &CampaignProcessor{
		templateUC:   templateUC,
		templateRend: templateRend,
		queueRepo:    queueRepo,
		accountRepo:  accountRepo,
		transactor:   transactor,
		ingestUC:     ingestUC,
		logger:       logger,
	}
}

func (p *CampaignProcessor) ProcessCampaign(ctx context.Context, campaign *entity.NotificationCampaign) error {
	tmpl, err := p.templateUC.GetTemplate(ctx, campaign.TemplateID)
	if err != nil {
		return fmt.Errorf("failed to load campaign template %s: %w", campaign.TemplateID, err)
	}

	accountIDs, err := p.resolveRecipients(ctx, campaign)
	if err != nil {
		return fmt.Errorf("failed to resolve campaign recipients: %w", err)
	}

	channels := p.resolveChannels(&tmpl.DefaultContent, campaign.CustomContent)

	content := tmpl.DefaultContent
	if campaign.CustomContent != nil {
		content = *campaign.CustomContent
	}

	for _, accountID := range accountIDs {
		if err := p.enqueueForRecipient(ctx, campaign, tmpl, accountID, channels, content); err != nil {
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
		return p.accountRepo.FindAll(ctx)
	case entity.CampaignTypeSegmented:
		if len(segment) == 0 {
			return nil, fmt.Errorf("segment filters required for segmented campaign")
		}
		return p.accountRepo.FindBySegment(ctx, segment)
	default:
		return nil, fmt.Errorf("unknown campaign type: %s", campaignType)
	}
}

func (p *CampaignProcessor) resolveRecipients(ctx context.Context, campaign *entity.NotificationCampaign) ([]uuid.UUID, error) {
	var segment map[string]interface{}
	if campaign.TargetSegment != nil {
		segment = *campaign.TargetSegment
	}

	accountIDs, err := p.ResolveSegment(ctx, campaign.CampaignType, segment)
	if err != nil {
		return nil, err
	}

	if len(accountIDs) == 0 {
		p.logger.Warn("Campaign resolved to zero recipients",
			core.String("campaignID", campaign.ID.String()),
		)
	}

	return accountIDs, nil
}

func (p *CampaignProcessor) resolveChannels(defaultContent, customContent *datatypes.JSONMap) []entity.Channel {
	source := defaultContent
	if customContent != nil {
		source = customContent
	}

	var channels []entity.Channel
	if source == nil {
		return channels
	}
	for key := range *source {
		ch := entity.Channel(key)
		switch ch {
		case entity.ChannelEmail, entity.ChannelInApp, entity.ChannelPush, entity.ChannelSMS:
			channels = append(channels, ch)
		}
	}
	return channels
}

func (p *CampaignProcessor) enqueueForRecipient(
	ctx context.Context,
	campaign *entity.NotificationCampaign,
	tmpl *entity.NotificationTemplate,
	accountID uuid.UUID,
	channels []entity.Channel,
	content datatypes.JSONMap,
) error {
	subject := ""
	if campaign.CustomSubject != nil {
		subject = *campaign.CustomSubject
	} else if s, ok := content["email"].(map[string]interface{}); ok {
		if subj, ok := s["subject"].(string); ok {
			subject = subj
		}
	}

	rendered, err := p.templateRend.Render(content, nil)
	if err != nil {
		return fmt.Errorf("failed to render content for account %s: %w", accountID, err)
	}

	for _, ch := range channels {
		channelContent, hasChannel := rendered[string(ch)]
		if !hasChannel {
			continue
		}

		channelMap, ok := channelContent.(map[string]interface{})
		if !ok {
			p.logger.Warn("Channel content is not a map, skipping",
				core.String("channel", string(ch)),
				core.String("accountID", accountID.String()),
			)
			continue
		}

		payload := datatypes.JSONMap{
			"title":     channelMap["title"],
			"content":   channelMap["body"],
			"actionUrl": channelMap["actionUrl"],
		}

		if ch == entity.ChannelEmail {
			payload["to"] = ""
			payload["subject"] = subject
		}

		now := time.Now().UTC()
		queueItem := &entity.NotificationQueue{
			NotificationType: tmpl.NotificationType,
			AccountID:        accountID,
			Priority:         tmpl.Priority,
			TemplateID:       &tmpl.ID,
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
