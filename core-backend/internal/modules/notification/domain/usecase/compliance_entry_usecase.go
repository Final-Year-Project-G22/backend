package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type ComplianceEntryUsecase interface {
	Create(ctx context.Context, accountID uuid.UUID, input CreateComplianceEntryInput) (*entity.ComplianceEntry, error)
	ListByBusinessProfile(ctx context.Context, accountID uuid.UUID, businessProfileID uuid.UUID) ([]*entity.ComplianceEntry, error)
	Update(ctx context.Context, accountID uuid.UUID, id uuid.UUID, input UpdateComplianceEntryInput) error
	Delete(ctx context.Context, accountID uuid.UUID, id uuid.UUID) error
	GetCalendar(ctx context.Context, accountID uuid.UUID) (*ComplianceCalendar, error)
}
