package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

type complianceEntryUsecase struct {
	complianceRepo notifrepo.ComplianceEntryRepository
	scheduledRepo  notifrepo.UserScheduledNotificationRepository
	profileReader  notifrepo.AccountReader
}

func NewComplianceEntryUsecase(
	complianceRepo notifrepo.ComplianceEntryRepository,
	scheduledRepo notifrepo.UserScheduledNotificationRepository,
	profileReader notifrepo.AccountReader,
) usecase.ComplianceEntryUsecase {
	return &complianceEntryUsecase{
		complianceRepo: complianceRepo,
		scheduledRepo:  scheduledRepo,
		profileReader:  profileReader,
	}
}

func (uc *complianceEntryUsecase) Create(ctx context.Context, accountID uuid.UUID, input usecase.CreateComplianceEntryInput) (*entity.ComplianceEntry, error) {
	if input.ExpiryDate.Before(time.Now().UTC()) {
		return nil, errors.New("notification: expiry date must be in the future")
	}
	if input.ReminderDaysBefore <= 0 {
		input.ReminderDaysBefore = 30
	}

	entry := &entity.ComplianceEntry{
		BusinessProfileID:  input.BusinessProfileID,
		AccountID:          accountID,
		ComplianceType:     input.ComplianceType,
		ReferenceNumber:    input.ReferenceNumber,
		IssuedDate:         input.IssuedDate,
		ExpiryDate:         input.ExpiryDate,
		ReminderDaysBefore: input.ReminderDaysBefore,
		Status:             entity.ComplianceEntryStatusActive,
	}

	if err := uc.complianceRepo.Create(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (uc *complianceEntryUsecase) ListByBusinessProfile(ctx context.Context, accountID uuid.UUID, businessProfileID uuid.UUID) ([]*entity.ComplianceEntry, error) {
	return uc.complianceRepo.ListByBusinessProfile(ctx, businessProfileID)
}

func (uc *complianceEntryUsecase) Update(ctx context.Context, accountID uuid.UUID, id uuid.UUID, input usecase.UpdateComplianceEntryInput) error {
	entry, err := uc.complianceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if entry == nil {
		return notiferror.ErrComplianceEntryNotFound
	}

	updates := map[string]interface{}{}
	if input.ReferenceNumber != nil {
		updates["reference_number"] = *input.ReferenceNumber
	}
	if input.IssuedDate != nil {
		updates["issued_date"] = *input.IssuedDate
	}
	if input.ExpiryDate != nil {
		updates["expiry_date"] = *input.ExpiryDate
	}
	if input.ReminderDaysBefore != nil {
		updates["reminder_days_before"] = *input.ReminderDaysBefore
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now().UTC()

	return uc.complianceRepo.UpdateByID(ctx, id, updates)
}

func (uc *complianceEntryUsecase) Delete(ctx context.Context, accountID uuid.UUID, id uuid.UUID) error {
	return uc.complianceRepo.Delete(ctx, id)
}

func (uc *complianceEntryUsecase) GetCalendar(ctx context.Context, accountID uuid.UUID) (*usecase.ComplianceCalendar, error) {
	now := time.Now().UTC()
	var entries []usecase.CalendarEntry

	complianceEntries, err := uc.complianceRepo.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	for _, ce := range complianceEntries {
		if ce.ExpiryDate.Before(now) {
			continue
		}
		daysRemaining := int(ce.ExpiryDate.Sub(now).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
		entries = append(entries, usecase.CalendarEntry{
			ID:              ce.ID,
			Type:            "compliance",
			Title:           string(ce.ComplianceType),
			ReferenceNumber: ce.ReferenceNumber,
			Date:            ce.ExpiryDate,
			DaysRemaining:   daysRemaining,
			Status:          string(ce.Status),
		})
	}

	scheduledAlerts, err := uc.scheduledRepo.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	for _, sa := range scheduledAlerts {
		daysRemaining := int(sa.ScheduledFor.Sub(now).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
		entries = append(entries, usecase.CalendarEntry{
			ID:            sa.ID,
			Type:          "scheduled_alert",
			Title:         sa.Title,
			Date:          sa.ScheduledFor,
			DaysRemaining: daysRemaining,
			Status:        string(sa.Status),
		})
	}

	return &usecase.ComplianceCalendar{Entries: entries}, nil
}
