package taxonomy

import (
	"context"

	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type TaxonomyValidator struct {
	sectorRepo iamrepo.SectorRepository
	tagRepo    iamrepo.TagRepository
}

func NewTaxonomyValidator(
	sectorRepo iamrepo.SectorRepository,
	tagRepo iamrepo.TagRepository,
) *TaxonomyValidator {
	return &TaxonomyValidator{
		sectorRepo: sectorRepo,
		tagRepo:    tagRepo,
	}
}

func (v *TaxonomyValidator) Validate(ctx context.Context, sectorIDs, tagIDs []uuid.UUID) error {
	if len(sectorIDs) > 0 {
		result, err := v.sectorRepo.FindByIDs(ctx, sectorIDs)
		if err != nil {
			return err
		}
		if len(result) != len(sectorIDs) {
			return errors.BadRequestError("errors.invalidInput")
		}
	}

	if len(tagIDs) > 0 {
		result, err := v.tagRepo.FindByIDs(ctx, tagIDs)
		if err != nil {
			return err
		}
		if len(result) != len(tagIDs) {
			return errors.BadRequestError("errors.invalidInput")
		}
	}

	return nil
}
