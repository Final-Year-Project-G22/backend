package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifevent "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	"github.com/google/uuid"
)

const businessAlertPollInterval = 1 * time.Hour
const businessAlertBatchSize = 100

type BusinessAlertScheduler struct {
	complianceRepo     notifrepo.ComplianceEntryRepository
	notifOutboxRepo    notifrepo.NotificationOutboxRepository
	accountReader      notifrepo.AccountReader
	complianceTypeRepo notifrepo.ComplianceTypeRepository
	logger             core.Logger
}

func NewBusinessAlertScheduler(
	complianceRepo notifrepo.ComplianceEntryRepository,
	notifOutboxRepo notifrepo.NotificationOutboxRepository,
	accountReader notifrepo.AccountReader,
	complianceTypeRepo notifrepo.ComplianceTypeRepository,
	logger core.Logger,
) *BusinessAlertScheduler {
	return &BusinessAlertScheduler{
		complianceRepo:     complianceRepo,
		notifOutboxRepo:    notifOutboxRepo,
		accountReader:      accountReader,
		complianceTypeRepo: complianceTypeRepo,
		logger:             logger,
	}
}

func (s *BusinessAlertScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *BusinessAlertScheduler) run(ctx context.Context) {
	// Run once immediately on startup
	s.processExpiring(ctx)
	ticker := time.NewTicker(businessAlertPollInterval)
	defer ticker.Stop()
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
		if strings.Contains(err.Error(), "does not exist") {
			return
		}
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
	if entry.BusinessProfileID == uuid.Nil {
		s.logger.Warn("Compliance entry has no business profile ID, skipping",
			core.String("complianceID", entry.ID.String()))
		return nil
	}

	accountInfo, err := s.accountReader.GetAccountInfo(ctx, entry.AccountID)
	if err != nil {
		return err
	}

	daysRemaining := int(entry.ExpiryDate.Sub(time.Now().UTC()).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	locale := accountInfo.Locale
	if locale == "" {
		locale = "en"
	}

	var expiryDate string
	if strings.HasPrefix(locale, "am") {
		amharicMonths := map[time.Month]string{
			time.January: "ጥር", time.February: "የካቲት", time.March: "መጋቢት",
			time.April: "ሚያዝያ", time.May: "ግንቦት", time.June: "ሰኔ",
			time.July: "ሐምሌ", time.August: "ነሐሴ", time.September: "መስከረም",
			time.October: "ጥቅምት", time.November: "ኅዳር", time.December: "ታኅሣሥ",
		}
		expiryDate = fmt.Sprintf("%s %d, %d", amharicMonths[entry.ExpiryDate.Month()], entry.ExpiryDate.Day(), entry.ExpiryDate.Year())
	} else {
		expiryDate = entry.ExpiryDate.Format("January 2, 2006")
	}

	complianceLabel, _ := s.complianceTypeRepo.GetLabel(ctx, string(entry.ComplianceType), locale)
	variables := map[string]string{
		"complianceType": complianceLabel,
		"daysRemaining":  fmt.Sprintf("%d", daysRemaining),
		"expiryDate":     expiryDate,
	}

	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        notifevent.ComplianceAlert,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "notification",
		AccountID:        entry.AccountID,
		NotificationType: string(entity.NotificationTypeComplianceInfo),
		ChannelPolicy:    notificationevent.ChannelPolicyAllEnabled,
		Variables:        variables,
		Metadata: notificationevent.Metadata{
			IdempotencyKey: "compliance-alert:" + entry.ID.String() + ":" + uuid.New().String(),
			Locale:         &locale,
		},
	}

	payload := make(map[string]interface{})
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	outbox := &entity.NotificationOutbox{
		EventType:      env.EventType,
		SchemaVersion:  env.SchemaVersion,
		SourceModule:   env.SourceModule,
		AccountID:      env.AccountID,
		IdempotencyKey: env.Metadata.IdempotencyKey,
		Payload:        payload,
		Status:         entity.NotificationOutboxStatusPending,
	}
	return s.notifOutboxRepo.Create(ctx, outbox)
}
