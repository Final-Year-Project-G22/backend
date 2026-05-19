package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const businessAlertPollInterval = 1 * time.Hour
const businessAlertBatchSize = 100

type BusinessAlertScheduler struct {
	complianceRepo notifrepo.ComplianceEntryRepository
	queueRepo      notifrepo.NotificationQueueRepository
	accountReader  notifrepo.AccountReader
	logger         core.Logger
}

func NewBusinessAlertScheduler(
	complianceRepo notifrepo.ComplianceEntryRepository,
	queueRepo notifrepo.NotificationQueueRepository,
	accountReader notifrepo.AccountReader,
	logger core.Logger,
) *BusinessAlertScheduler {
	return &BusinessAlertScheduler{
		complianceRepo: complianceRepo,
		queueRepo:      queueRepo,
		accountReader:  accountReader,
		logger:         logger,
	}
}

func (s *BusinessAlertScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *BusinessAlertScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(businessAlertPollInterval)
	defer ticker.Stop()
	// Run once immediately on startup
	s.processExpiring(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processExpiring(ctx)
		}
	}
}

func (s *BusinessAlertScheduler) processExpiring(ctx context.Context) {
	now := time.Now().UTC()
	entries, err := s.complianceRepo.FetchExpiringSoon(ctx, now, businessAlertBatchSize)
	if err != nil {
		s.logger.Error("Failed to fetch expiring compliance entries", core.Error(err))
		return
	}

	for _, entry := range entries {
		if err := s.sendAlert(ctx, entry); err != nil {
			s.logger.Error("Failed to send business alert",
				core.String("complianceID", entry.ID.String()),
				core.Error(err),
			)
			continue
		}

		updates := map[string]interface{}{
			"last_notified_at": now,
			"updated_at":       now,
		}
		if err := s.complianceRepo.UpdateByID(ctx, entry.ID, updates); err != nil {
			s.logger.Error("Failed to update compliance entry last_notified_at",
				core.String("complianceID", entry.ID.String()),
				core.Error(err),
			)
		}
	}
}

func (s *BusinessAlertScheduler) sendAlert(ctx context.Context, entry *entity.ComplianceEntry) error {
	// Determine the target account from the compliance entry
	// The compliance entry has a business_profile_id, but we don't have direct access
	// to the business profile repo here. We enqueue to the notification_queue and
	// the delivery pipeline will use the account info from the payload.

	// For now, we enqueue with in_app + email channels
	// The accountID is empty since we don't have a direct mapping from compliance entry
	// to account. This should be resolved by joining through business_profile.
	// We log a warning and skip if accountID can't be resolved.
	if entry.BusinessProfileID == uuid.Nil {
		s.logger.Warn("Compliance entry has no business profile ID, skipping",
			core.String("complianceID", entry.ID.String()))
		return nil
	}

	daysRemaining := int(entry.ExpiryDate.Sub(time.Now().UTC()).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	title := fmt.Sprintf("%s Expiring", entry.ComplianceType)
	body := fmt.Sprintf(
		"Your %s is expiring in %d days (expiry: %s). Please renew before the deadline.",
		entry.ComplianceType, daysRemaining, entry.ExpiryDate.Format("Jan 2, 2006"),
	)

	payload := datatypes.JSONMap{
		"title":               title,
		"body":                body,
		"compliance_entry_id": entry.ID.String(),
		"compliance_type":     string(entry.ComplianceType),
		"expiry_date":         entry.ExpiryDate.Format(time.RFC3339),
	}

	// Enqueue for both in_app and email
	for _, channel := range []entity.Channel{entity.ChannelInApp, entity.ChannelEmail} {
		queueItem := &entity.NotificationQueue{
			NotificationType: entity.NotificationTypeAccountAlertInfo,
			AccountID:        entry.AccountID,
			Priority:         entity.NotificationPriorityMedium,
			Channel:          channel,
			Payload:          payload,
			ScheduledFor:     time.Now().UTC(),
			MaxRetries:       3,
			RetryCount:       0,
			Status:           entity.NotificationStatusPending,
		}
		if err := s.queueRepo.Create(ctx, queueItem); err != nil {
			return fmt.Errorf("failed to enqueue business alert: %w", err)
		}
	}

	return nil
}
