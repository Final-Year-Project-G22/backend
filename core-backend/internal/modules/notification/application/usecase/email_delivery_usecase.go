package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

type emailDeliveryUsecase struct {
	deliveryLogRepo repository.EmailDeliveryLogRepository
	historyRepo     repository.NotificationHistoryRepository
	logger          core.Logger
}

func NewEmailDeliveryUsecase(
	deliveryLogRepo repository.EmailDeliveryLogRepository,
	historyRepo repository.NotificationHistoryRepository,
	logger core.Logger,
) usecase.EmailDeliveryUsecase {
	return &emailDeliveryUsecase{
		deliveryLogRepo: deliveryLogRepo,
		historyRepo:     historyRepo,
		logger:          logger,
	}
}

func (uc *emailDeliveryUsecase) HandleWebhookEvent(ctx context.Context, event usecase.ResendWebhookEvent) error {
	deliveryLog, err := uc.deliveryLogRepo.GetByProviderMessageID(ctx, event.EmailID)
	if err != nil {
		uc.logger.Error("Delivery log not found for webhook event",
			core.String("emailID", event.EmailID),
			core.String("eventType", event.EventType),
		)
		return err
	}

	metadata := make(map[string]interface{})
	if event.BounceReason != nil {
		metadata["bounceReason"] = *event.BounceReason
	}

	if err := uc.deliveryLogRepo.UpdateDeliveryEvent(ctx, deliveryLog.ID, event.EventType, event.OccurredAt, metadata); err != nil {
		uc.logger.Error("Failed to update delivery event",
			core.String("emailID", event.EmailID),
			core.String("eventType", event.EventType),
			core.Error(err),
		)
		return err
	}

	switch event.EventType {
	case "opened":
		if herr := uc.historyRepo.MarkRead(ctx, deliveryLog.NotificationHistoryID); herr != nil {
			uc.logger.Warn("Failed to mark notification history as read on open",
				core.String("historyID", deliveryLog.NotificationHistoryID.String()),
				core.Error(herr),
			)
		}
	case "clicked":
		if herr := uc.historyRepo.MarkClicked(ctx, deliveryLog.NotificationHistoryID); herr != nil {
			uc.logger.Warn("Failed to mark notification history as clicked",
				core.String("historyID", deliveryLog.NotificationHistoryID.String()),
				core.Error(herr),
			)
		}
	}

	return nil
}

func (uc *emailDeliveryUsecase) GetDeliveryLog(ctx context.Context, historyID uuid.UUID) (*entity.EmailDeliveryLog, error) {
	return uc.deliveryLogRepo.GetByNotificationHistoryID(ctx, historyID)
}

func (uc *emailDeliveryUsecase) GetDeliveryLogByProviderID(ctx context.Context, providerMessageID string) (*entity.EmailDeliveryLog, error) {
	return uc.deliveryLogRepo.GetByProviderMessageID(ctx, providerMessageID)
}
