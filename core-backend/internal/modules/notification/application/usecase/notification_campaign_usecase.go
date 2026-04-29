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
	campaignRepo repository.NotificationCampaignRepository
	templateRepo repository.NotificationTemplateRepository
	queueRepo    repository.NotificationQueueRepository
	campaignProc *service.CampaignProcessor
	transactor   sharedrepo.Transactor
	logger       core.Logger
}

func NewNotificationCampaignUsecase(
	campaignRepo repository.NotificationCampaignRepository,
	templateRepo repository.NotificationTemplateRepository,
	queueRepo repository.NotificationQueueRepository,
	campaignProc *service.CampaignProcessor,
	transactor sharedrepo.Transactor,
	logger core.Logger,
) usecase.NotificationCampaignUsecase {
	return &notificationCampaignUsecase{
		campaignRepo: campaignRepo,
		templateRepo: templateRepo,
		queueRepo:    queueRepo,
		campaignProc: campaignProc,
		transactor:   transactor,
		logger:       logger,
	}
}

func (uc *notificationCampaignUsecase) CreateCampaign(ctx context.Context, createdBy uuid.UUID, input usecase.CreateCampaignInput) (*entity.NotificationCampaign, error) {
	if _, err := uc.templateRepo.GetByID(ctx, input.TemplateID); err != nil {
		return nil, notiferror.ErrTemplateNotFound
	}

	if input.CampaignType == entity.CampaignTypeSegmented && input.TargetSegment == nil {
		return nil, fmt.Errorf("target segment required for segmented campaign")
	}

	var targetSegment *datatypes.JSONMap
	if input.TargetSegment != nil {
		ts := datatypes.JSONMap(*input.TargetSegment)
		targetSegment = &ts
	}

	var customContent *datatypes.JSONMap
	if input.CustomContent != nil {
		cc := datatypes.JSONMap(*input.CustomContent)
		customContent = &cc
	}

	campaign := &entity.NotificationCampaign{
		Name:          input.Name,
		Description:   input.Description,
		CampaignType:  input.CampaignType,
		TargetSegment: targetSegment,
		TemplateID:    input.TemplateID,
		CustomSubject: input.CustomSubject,
		CustomContent: customContent,
		ScheduledFor:  input.ScheduledFor,
		Status:        entity.CampaignStatusDraft,
		CreatedBy:     createdBy,
	}

	if err := uc.campaignRepo.Create(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

func (uc *notificationCampaignUsecase) GetCampaign(ctx context.Context, id uuid.UUID) (*entity.NotificationCampaign, error) {
	return uc.campaignRepo.GetByID(ctx, id)
}

func (uc *notificationCampaignUsecase) ListCampaigns(ctx context.Context, status *entity.CampaignStatus, q query.QueryOptions) ([]*entity.NotificationCampaign, error) {
	if status != nil {
		return uc.campaignRepo.ListByStatus(ctx, *status, q)
	}
	return uc.campaignRepo.Find(ctx, q)
}

func (uc *notificationCampaignUsecase) UpdateCampaign(ctx context.Context, id uuid.UUID, input usecase.UpdateCampaignInput) (*entity.NotificationCampaign, error) {
	campaign, err := uc.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if campaign.Status != entity.CampaignStatusDraft {
		return nil, notiferror.ErrCampaignNotDraft
	}

	if input.Name != nil {
		campaign.Name = *input.Name
	}
	if input.Description != nil {
		campaign.Description = input.Description
	}
	if input.TargetSegment != nil {
		ts := datatypes.JSONMap(*input.TargetSegment)
		campaign.TargetSegment = &ts
	}
	if input.CustomSubject != nil {
		campaign.CustomSubject = input.CustomSubject
	}
	if input.CustomContent != nil {
		cc := datatypes.JSONMap(*input.CustomContent)
		campaign.CustomContent = &cc
	}
	if input.ScheduledFor != nil {
		campaign.ScheduledFor = input.ScheduledFor
	}

	if err := uc.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

func (uc *notificationCampaignUsecase) ScheduleCampaign(ctx context.Context, input usecase.ScheduleCampaignInput) error {
	campaign, err := uc.campaignRepo.GetByID(ctx, input.CampaignID)
	if err != nil {
		return err
	}

	if campaign.Status != entity.CampaignStatusDraft {
		return notiferror.ErrCampaignNotDraft
	}

	if _, err := uc.templateRepo.GetByID(ctx, campaign.TemplateID); err != nil {
		return notiferror.ErrTemplateNotFound
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

	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.campaignRepo.UpdateByID(txCtx, campaign.ID, map[string]interface{}{
			"target_segment": resolvedMap,
			"status":         entity.CampaignStatusScheduled,
			"scheduled_for":  scheduledFor,
		}); err != nil {
			return err
		}
		return nil
	})
}

func (uc *notificationCampaignUsecase) CancelCampaign(ctx context.Context, id uuid.UUID) error {
	campaign, err := uc.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if campaign.Status != entity.CampaignStatusScheduled && campaign.Status != entity.CampaignStatusSending {
		return notiferror.ErrCampaignInvalidTransition
	}

	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.campaignRepo.UpdateStatus(txCtx, id, entity.CampaignStatusCancelled); err != nil {
			return err
		}
		if err := uc.queueRepo.CancelByCampaign(txCtx, id); err != nil {
			return err
		}
		return nil
	})
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

	if err := uc.campaignProc.ProcessCampaign(ctx, campaign); err != nil {
		return err
	}

	return uc.campaignRepo.UpdateStatus(ctx, campaign.ID, entity.CampaignStatusCompleted)
}
