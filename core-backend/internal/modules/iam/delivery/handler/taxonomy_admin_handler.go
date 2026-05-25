package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
)

type TaxonomyAdminHandler struct {
	sectorRepo repository.SectorRepository
	tagRepo    repository.TagRepository
}

func NewTaxonomyAdminHandler(sectorRepo repository.SectorRepository, tagRepo repository.TagRepository) *TaxonomyAdminHandler {
	return &TaxonomyAdminHandler{sectorRepo: sectorRepo, tagRepo: tagRepo}
}

// --- Sector ---

func (h *TaxonomyAdminHandler) HandleListSectors(ctx context.Context, input *dto.ListSectorsInput) (*dto.ListSectorsOutput, error) {
	q := query.QueryOptions{Preload: []string{"Translations"}}
	if input.Page > 0 {
		q.Page = input.Page
	}
	if input.PageSize > 0 {
		q.PageSize = input.PageSize
	}
	q.Search = input.Search

	result := h.sectorRepo.FindAll(ctx, q)
	items := make([]dto.SectorResponse, 0, len(result.Data))
	for _, s := range result.Data {
		items = append(items, sectorToResponse(s))
	}
	return &dto.ListSectorsOutput{Body: dto.ListSectorsResponseBody{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}}, nil
}

func (h *TaxonomyAdminHandler) HandleGetSector(ctx context.Context, input *dto.GetSectorInput) (*dto.GetSectorOutput, error) {
	sector, err := h.sectorRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetSectorOutput{Body: sectorToResponse(sector)}, nil
}

func (h *TaxonomyAdminHandler) HandleCreateSector(ctx context.Context, input *dto.CreateSectorInput) (*dto.CreateSectorOutput, error) {
	body := input.Body
	sector := &entity.Sector{
		Slug:     body.Slug,
		ParentID: body.ParentID,
		Translations: []entity.SectorTranslation{
			{Language: constants.LocaleEnglish, Name: body.NameEN, Description: body.DescEN},
			{Language: constants.LocaleAmharic, Name: body.NameAM, Description: body.DescAM},
		},
	}
	if err := h.sectorRepo.Create(ctx, sector); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateSectorOutput{Body: dto.CreateSectorResponseBody{ID: sector.ID}}, nil
}

func (h *TaxonomyAdminHandler) HandleUpdateSector(ctx context.Context, input *dto.UpdateSectorInput) (*dto.UpdateSectorOutput, error) {
	sector, err := h.sectorRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	body := input.Body
	if body.Slug != nil {
		sector.Slug = *body.Slug
	}
	if body.ParentID != nil {
		sector.ParentID = body.ParentID
	}
	if err := h.sectorRepo.Update(ctx, sector); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	if body.NameEN != nil || body.NameAM != nil {
		if body.NameEN != nil {
			if err := h.sectorRepo.UpsertTranslation(ctx, &entity.SectorTranslation{
				SectorID:    sector.ID,
				Language:    constants.LocaleEnglish,
				Name:        *body.NameEN,
				Description: body.DescEN,
			}); err != nil {
				return nil, apperrors.ToHumaError(ctx, err)
			}
		}
		if body.NameAM != nil {
			if err := h.sectorRepo.UpsertTranslation(ctx, &entity.SectorTranslation{
				SectorID:    sector.ID,
				Language:    constants.LocaleAmharic,
				Name:        *body.NameAM,
				Description: body.DescAM,
			}); err != nil {
				return nil, apperrors.ToHumaError(ctx, err)
			}
		}
	}
	return &dto.UpdateSectorOutput{Body: dto.UpdateSectorResponseBody{Message: i18n.T(ctx, "iam.successes.sectorUpdated")}}, nil
}

func (h *TaxonomyAdminHandler) HandleDeleteSector(ctx context.Context, input *dto.GetSectorInput) (*dto.UpdateSectorOutput, error) {
	if err := h.sectorRepo.Delete(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateSectorOutput{Body: dto.UpdateSectorResponseBody{Message: i18n.T(ctx, "iam.successes.sectorDeleted")}}, nil
}

// --- Tag ---

func (h *TaxonomyAdminHandler) HandleListTags(ctx context.Context, input *dto.ListTagsInput) (*dto.ListTagsOutput, error) {
	q := query.QueryOptions{Preload: []string{"Translations"}}
	if input.Page > 0 {
		q.Page = input.Page
	}
	if input.PageSize > 0 {
		q.PageSize = input.PageSize
	}
	q.Search = input.Search

	result := h.tagRepo.FindAll(ctx, q)
	items := make([]dto.TagResponse, 0, len(result.Data))
	for _, t := range result.Data {
		items = append(items, tagToResponse(t))
	}
	return &dto.ListTagsOutput{Body: dto.ListTagsResponseBody{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}}, nil
}

func (h *TaxonomyAdminHandler) HandleGetTag(ctx context.Context, input *dto.GetTagInput) (*dto.GetTagOutput, error) {
	tag, err := h.tagRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetTagOutput{Body: tagToResponse(tag)}, nil
}

func (h *TaxonomyAdminHandler) HandleCreateTag(ctx context.Context, input *dto.CreateTagInput) (*dto.CreateTagOutput, error) {
	body := input.Body
	tag := &entity.Tag{
		Slug:          body.Slug,
		Group:         body.Group,
		IsMultiSelect: body.IsMultiSelect,
		Translations: []entity.TagTranslation{
			{Language: constants.LocaleEnglish, Name: body.NameEN, Description: body.DescEN},
			{Language: constants.LocaleAmharic, Name: body.NameAM, Description: body.DescAM},
		},
	}
	if err := h.tagRepo.Create(ctx, tag); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateTagOutput{Body: dto.CreateTagResponseBody{ID: tag.ID}}, nil
}

func (h *TaxonomyAdminHandler) HandleUpdateTag(ctx context.Context, input *dto.UpdateTagInput) (*dto.UpdateTagOutput, error) {
	tag, err := h.tagRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	body := input.Body
	if body.Slug != nil {
		tag.Slug = *body.Slug
	}
	if body.Group != nil {
		tag.Group = *body.Group
	}
	if body.IsMultiSelect != nil {
		tag.IsMultiSelect = *body.IsMultiSelect
	}
	if err := h.tagRepo.Update(ctx, tag); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	if body.NameEN != nil || body.NameAM != nil {
		if body.NameEN != nil {
			if err := h.tagRepo.UpsertTranslation(ctx, &entity.TagTranslation{
				TagID:       tag.ID,
				Language:    constants.LocaleEnglish,
				Name:        *body.NameEN,
				Description: body.DescEN,
			}); err != nil {
				return nil, apperrors.ToHumaError(ctx, err)
			}
		}
		if body.NameAM != nil {
			if err := h.tagRepo.UpsertTranslation(ctx, &entity.TagTranslation{
				TagID:       tag.ID,
				Language:    constants.LocaleAmharic,
				Name:        *body.NameAM,
				Description: body.DescAM,
			}); err != nil {
				return nil, apperrors.ToHumaError(ctx, err)
			}
		}
	}
	return &dto.UpdateTagOutput{Body: dto.UpdateTagResponseBody{Message: i18n.T(ctx, "iam.successes.tagUpdated")}}, nil
}

func (h *TaxonomyAdminHandler) HandleDeleteTag(ctx context.Context, input *dto.GetTagInput) (*dto.UpdateTagOutput, error) {
	if err := h.tagRepo.Delete(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateTagOutput{Body: dto.UpdateTagResponseBody{Message: i18n.T(ctx, "iam.successes.tagDeleted")}}, nil
}

// --- Mappers ---

func sectorToResponse(s *entity.Sector) dto.SectorResponse {
	var nameEN, nameAM string
	var descEN, descAM *string
	for _, t := range s.Translations {
		switch t.Language {
		case constants.LocaleEnglish:
			nameEN = t.Name
			descEN = t.Description
		case constants.LocaleAmharic:
			nameAM = t.Name
			descAM = t.Description
		}
	}
	return dto.SectorResponse{
		ID:        s.ID,
		Slug:      s.Slug,
		ParentID:  s.ParentID,
		NameEN:    nameEN,
		DescEN:    descEN,
		NameAM:    nameAM,
		DescAM:    descAM,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func tagToResponse(t *entity.Tag) dto.TagResponse {
	var nameEN, nameAM string
	var descEN, descAM *string
	for _, tr := range t.Translations {
		switch tr.Language {
		case constants.LocaleEnglish:
			nameEN = tr.Name
			descEN = tr.Description
		case constants.LocaleAmharic:
			nameAM = tr.Name
			descAM = tr.Description
		}
	}
	return dto.TagResponse{
		ID:            t.ID,
		Slug:          t.Slug,
		Group:         t.Group,
		IsMultiSelect: t.IsMultiSelect,
		NameEN:        nameEN,
		DescEN:        descEN,
		NameAM:        nameAM,
		DescAM:        descAM,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}
