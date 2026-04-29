package service

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
)

const campaignPollInterval = 30 * time.Second

type CampaignScheduler struct {
	campaignUC usecase.NotificationCampaignUsecase
}

func NewCampaignScheduler(campaignUC usecase.NotificationCampaignUsecase) *CampaignScheduler {
	return &CampaignScheduler{campaignUC: campaignUC}
}

func (s *CampaignScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *CampaignScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(campaignPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.campaignUC.ProcessScheduledCampaigns(ctx)
		}
	}
}
