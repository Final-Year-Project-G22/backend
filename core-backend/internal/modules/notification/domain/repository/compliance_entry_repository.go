package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type ComplianceEntryRepository interface {
	sharedrepo.GenericRepository[entity.ComplianceEntry]

	ListByBusinessProfile(ctx context.Context, businessProfileID uuid.UUID) ([]*entity.ComplianceEntry, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.ComplianceEntry, error)
	FetchExpiringSoon(ctx context.Context, now time.Time, limit int) ([]*entity.ComplianceEntry, error)
	CountByStatus(ctx context.Context, businessProfileID uuid.UUID, status entity.ComplianceEntryStatus) (int64, error)
}
