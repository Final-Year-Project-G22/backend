package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type ComplianceTypeWithLabel struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
}

type ComplianceTypeRepository interface {
	sharedrepo.GenericRepository[entity.ComplianceTypeLocalization]

	GetLabel(ctx context.Context, complianceType string, locale string) (string, error)
	ListWithLabels(ctx context.Context, locale string) ([]ComplianceTypeWithLabel, error)
}
