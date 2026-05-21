package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"gorm.io/gorm"
)

type complianceTypeRepository struct {
	sharedrepo.GenericRepository[entity.ComplianceTypeLocalization]
	db     *core.Database
	logger core.Logger
}

func NewComplianceTypeRepository(db *core.Database, logger core.Logger) notifrepo.ComplianceTypeRepository {
	base := sharedrepo.NewBaseRepository[entity.ComplianceTypeLocalization](db, logger)
	return &complianceTypeRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *complianceTypeRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *complianceTypeRepository) GetLabel(ctx context.Context, complianceType string, locale string) (string, error) {
	var loc entity.ComplianceTypeLocalization
	if err := r.getDB(ctx).Where("compliance_type = ? AND locale = ?", complianceType, locale).First(&loc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if locale != "en" {
				return r.GetLabel(ctx, complianceType, "en")
			}
			return complianceType, nil
		}
		return "", errors.InternalError("errors.databaseError", err)
	}
	return loc.Label, nil
}

func (r *complianceTypeRepository) ListWithLabels(ctx context.Context, locale string) ([]notifrepo.ComplianceTypeWithLabel, error) {
	var locs []entity.ComplianceTypeLocalization
	if err := r.getDB(ctx).Where("locale = ?", locale).Find(&locs).Error; err != nil {
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if len(locs) == 0 && locale != "en" {
		return r.ListWithLabels(ctx, "en")
	}
	result := make([]notifrepo.ComplianceTypeWithLabel, 0, len(locs))
	for _, l := range locs {
		result = append(result, notifrepo.ComplianceTypeWithLabel{Slug: l.ComplianceType, Label: l.Label})
	}
	return result, nil
}
