package service

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/service"
	"github.com/google/uuid"
)

type SyncComplianceService struct {
	profileRepo    repository.BusinessProfileRepository
	complianceRepo notifrepo.ComplianceEntryRepository
	logger         core.Logger
}

func NewSyncComplianceService(
	profileRepo repository.BusinessProfileRepository,
	complianceRepo notifrepo.ComplianceEntryRepository,
	logger core.Logger,
) *SyncComplianceService {
	return &SyncComplianceService{
		profileRepo:    profileRepo,
		complianceRepo: complianceRepo,
		logger:         logger,
	}
}

func (s *SyncComplianceService) Sync(ctx context.Context, accountID uuid.UUID) error {
	profile, err := s.profileRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to fetch business profile for sync", core.String("accountID", accountID.String()), core.Error(err))
		return err
	}
	if profile == nil {
		return nil
	}

	fields := map[entity.ComplianceType]*string{
		entity.ComplianceTypeTaxRegistration:      profile.TaxIdentificationNumber,
		entity.ComplianceTypeTradeLicense:         profile.TradeLicenseNumber,
		entity.ComplianceTypeBusinessRegistration: profile.RegistrationNumber,
	}

	for ctype, value := range fields {
		if value != nil && *value != "" {
			s.upsert(ctx, accountID, profile.ID, ctype, *value)
		} else {
			s.expireIfAuto(ctx, accountID, ctype)
		}
	}

	return nil
}

func (s *SyncComplianceService) upsert(ctx context.Context, accountID, businessProfileID uuid.UUID, ctype entity.ComplianceType, referenceNumber string) {
	defaults := service.ComplianceDefaults[ctype]

	var existing bool
	entries, err := s.complianceRepo.ListByAccount(ctx, accountID)
	if err == nil {
		for _, e := range entries {
			if e.ComplianceType == ctype && e.Source == entity.ComplianceSourceAuto {
				existing = true
				if e.ReferenceNumber == nil || *e.ReferenceNumber != referenceNumber {
					updates := map[string]interface{}{
						"reference_number": referenceNumber,
						"updated_at":       time.Now().UTC(),
					}
					_ = s.complianceRepo.UpdateByID(ctx, e.ID, updates)
				}
				break
			}
		}
	}

	if !existing {
		expiry := time.Now().UTC().AddDate(0, 0, defaults.ExpiryDurationDays)
		entry := &entity.ComplianceEntry{
			BusinessProfileID:  businessProfileID,
			AccountID:          accountID,
			ComplianceType:     ctype,
			ReferenceNumber:    &referenceNumber,
			Source:             entity.ComplianceSourceAuto,
			ExpiryDate:         expiry,
			ReminderDaysBefore: defaults.ReminderDays,
			Status:             entity.ComplianceEntryStatusActive,
		}
		_ = s.complianceRepo.Create(ctx, entry)
	}
}

func (s *SyncComplianceService) expireIfAuto(ctx context.Context, accountID uuid.UUID, ctype entity.ComplianceType) {
	entries, err := s.complianceRepo.ListByAccount(ctx, accountID)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.ComplianceType == ctype && e.Source == entity.ComplianceSourceAuto && e.Status == entity.ComplianceEntryStatusActive {
			_ = s.complianceRepo.UpdateByID(ctx, e.ID, map[string]interface{}{
				"status":     entity.ComplianceEntryStatusExpired,
				"updated_at": time.Now().UTC(),
			})
			return
		}
	}
}
