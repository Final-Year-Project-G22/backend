package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type notificationCampaignUsecase struct {
	campaignRepo         repository.NotificationCampaignRepository
	campaignTemplateRepo repository.CampaignTemplateRepository
	queueRepo            repository.NotificationQueueRepository
	accountReader        repository.AccountReader
	campaignProc         *service.CampaignProcessor
	sseBroadcaster       *service.CampaignSSEBroadcaster
	transactor           sharedrepo.Transactor
	logger               core.Logger
}

func NewNotificationCampaignUsecase(
	campaignRepo repository.NotificationCampaignRepository,
	campaignTemplateRepo repository.CampaignTemplateRepository,
	queueRepo repository.NotificationQueueRepository,
	accountReader repository.AccountReader,
	campaignProc *service.CampaignProcessor,
	sseBroadcaster *service.CampaignSSEBroadcaster,
	transactor sharedrepo.Transactor,
	logger core.Logger,
) usecase.NotificationCampaignUsecase {
	return &notificationCampaignUsecase{
		campaignRepo:         campaignRepo,
		campaignTemplateRepo: campaignTemplateRepo,
		queueRepo:            queueRepo,
		accountReader:        accountReader,
		campaignProc:         campaignProc,
		sseBroadcaster:       sseBroadcaster,
		transactor:           transactor,
		logger:               logger,
	}
}

func (uc *notificationCampaignUsecase) CreateCampaign(ctx context.Context, createdBy uuid.UUID, input usecase.CreateCampaignInput) (*entity.NotificationCampaign, error) {
	if _, err := uc.campaignTemplateRepo.GetByID(ctx, input.CampaignTemplateID); err != nil {
		return nil, notiferror.ErrCampaignTemplateNotFound
	}

	if input.CampaignType == entity.CampaignTypeSegmented && input.TargetSegment == nil {
		return nil, fmt.Errorf("target segment required for segmented campaign")
	}

	var targetSegment *datatypes.JSONMap
	if input.TargetSegment != nil {
		ts := datatypes.JSONMap(*input.TargetSegment)
		targetSegment = &ts
	}

	campaign := &entity.NotificationCampaign{
		Name:               input.Name,
		Description:        input.Description,
		CampaignType:       input.CampaignType,
		TargetSegment:      targetSegment,
		CampaignTemplateID: input.CampaignTemplateID,
		ScheduledFor:       input.ScheduledFor,
		Status:             entity.CampaignStatusDraft,
		CreatedBy:          createdBy,
		SectorIDs:          input.SectorIDs,
		TagIDs:             input.TagIDs,
		Region:             input.Region,
		Stage:              input.Stage,
	}

	if err := uc.campaignRepo.Create(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

func (uc *notificationCampaignUsecase) GetCampaign(ctx context.Context, id uuid.UUID) (*usecase.CampaignDetail, error) {
	campaign, err := uc.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &usecase.CampaignDetail{
		Campaign: campaign,
	}

	tmpl, err := uc.campaignTemplateRepo.GetByID(ctx, campaign.CampaignTemplateID)
	if err == nil {
		detail.CampaignTemplate = tmpl
	}

	accountInfo, err := uc.accountReader.GetAccountInfo(ctx, campaign.CreatedBy)
	if err == nil {
		detail.CreatedByName = accountInfo.Name
		detail.CreatedByEmail = accountInfo.Email
	}

	return detail, nil
}

func (uc *notificationCampaignUsecase) ListCampaigns(ctx context.Context, status *entity.CampaignStatus, q query.QueryOptions) ([]*entity.NotificationCampaign, int64, error) {
	if status != nil {
		return uc.campaignRepo.ListByStatus(ctx, *status, q)
	}
	total, err := uc.campaignRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	campaigns, err := uc.campaignRepo.Find(ctx, q)
	return campaigns, total, err
}

func (uc *notificationCampaignUsecase) UpdateCampaign(ctx context.Context, id uuid.UUID, input usecase.UpdateCampaignInput) (*entity.NotificationCampaign, error) {
	campaign, err := uc.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if campaign.Status != entity.CampaignStatusDraft {
		return nil, notiferror.ErrCampaignNotDraft
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	if input.TargetSegment != nil {
		ts := datatypes.JSONMap(*input.TargetSegment)
		updates["target_segment"] = &ts
	}
	if input.ScheduledFor != nil {
		updates["scheduled_for"] = input.ScheduledFor
	}
	if input.SectorIDs != nil {
		updates["sector_ids"] = input.SectorIDs
	}
	if input.TagIDs != nil {
		updates["tag_ids"] = input.TagIDs
	}
	if input.Region != nil {
		updates["region"] = input.Region
	}
	if input.Stage != nil {
		updates["stage"] = input.Stage
	}

	if len(updates) > 0 {
		if err := uc.campaignRepo.UpdateByID(ctx, id, updates); err != nil {
			return nil, err
		}
	}

	return uc.campaignRepo.GetByID(ctx, id)
}

func (uc *notificationCampaignUsecase) ScheduleCampaign(ctx context.Context, input usecase.ScheduleCampaignInput) error {
	campaign, err := uc.campaignRepo.GetByID(ctx, input.CampaignID)
	if err != nil {
		return err
	}

	if campaign.Status != entity.CampaignStatusDraft {
		return notiferror.ErrCampaignNotDraft
	}

	if _, err := uc.campaignTemplateRepo.GetByID(ctx, campaign.CampaignTemplateID); err != nil {
		return notiferror.ErrCampaignTemplateNotFound
	}

	var segment map[string]interface{}
	if campaign.TargetSegment != nil {
		segment = *campaign.TargetSegment
	}

	accountIDs, err := uc.campaignProc.ResolveSegment(ctx, campaign.CampaignType, segment)
	if err != nil {
		return fmt.Errorf("failed to resolve campaign segment: %w", err)
	}

	resolved := make([]string, len(accountIDs))
	for i, id := range accountIDs {
		resolved[i] = id.String()
	}
	resolvedMap := datatypes.JSONMap{
		"resolvedAccountIDs": resolved,
		"totalCount":         len(accountIDs),
	}

	scheduledFor := time.Now().UTC()
	if campaign.ScheduledFor != nil && campaign.ScheduledFor.After(time.Now()) {
		scheduledFor = *campaign.ScheduledFor
	}

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.campaignRepo.UpdateByID(txCtx, campaign.ID, map[string]interface{}{
			"target_segment": resolvedMap,
			"status":         entity.CampaignStatusScheduled,
			"scheduled_for":  scheduledFor,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	uc.publishSSEEvent(campaign.ID, entity.CampaignStatusScheduled)
	return nil
}

func (uc *notificationCampaignUsecase) publishSSEEvent(campaignID uuid.UUID, status entity.CampaignStatus) {
	uc.sseBroadcaster.Publish(service.CampaignStatusEvent{
		CampaignID: campaignID,
		Status:     status,
		Timestamp:  time.Now().UTC(),
	})
}

func (uc *notificationCampaignUsecase) CancelCampaign(ctx context.Context, id uuid.UUID) error {
	campaign, err := uc.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if campaign.Status != entity.CampaignStatusScheduled {
		return notiferror.ErrCampaignInvalidTransition
	}

	if err := uc.campaignRepo.UpdateStatus(ctx, id, entity.CampaignStatusCancelled); err != nil {
		return err
	}
	uc.publishSSEEvent(id, entity.CampaignStatusCancelled)
	return nil
}

func (uc *notificationCampaignUsecase) ProcessScheduledCampaigns(ctx context.Context) error {
	campaigns, err := uc.campaignRepo.ListScheduled(ctx)
	if err != nil {
		return err
	}

	for _, campaign := range campaigns {
		if err := uc.processCampaign(ctx, campaign); err != nil {
			uc.logger.Error("Failed to process campaign",
				core.String("campaignID", campaign.ID.String()),
				core.Error(err),
			)
			continue
		}
	}

	return nil
}

func (uc *notificationCampaignUsecase) processCampaign(ctx context.Context, campaign *entity.NotificationCampaign) error {
	if err := uc.campaignRepo.UpdateStatus(ctx, campaign.ID, entity.CampaignStatusSending); err != nil {
		return err
	}
	uc.publishSSEEvent(campaign.ID, entity.CampaignStatusSending)

	if err := uc.campaignProc.ProcessCampaign(ctx, campaign); err != nil {
		return err
	}

	if err := uc.campaignRepo.UpdateStatus(ctx, campaign.ID, entity.CampaignStatusCompleted); err != nil {
		return err
	}
	uc.publishSSEEvent(campaign.ID, entity.CampaignStatusCompleted)
	return nil
}
