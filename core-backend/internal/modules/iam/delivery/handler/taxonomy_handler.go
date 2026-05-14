package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
)

type TaxonomyHandler struct {
	sectorRepo repository.SectorRepository
	tagRepo    repository.TagRepository
}

func NewTaxonomyHandler(sectorRepo repository.SectorRepository, tagRepo repository.TagRepository) *TaxonomyHandler {
	return &TaxonomyHandler{sectorRepo: sectorRepo, tagRepo: tagRepo}
}

func (h *TaxonomyHandler) HandleListSectors(ctx context.Context, input *dto.ListSectorsInput) (*dto.ListSectorsOutput, error) {
	q := query.QueryOptions{Preload: []string{"Translations"}}
	if input.Page > 0 {
		q.Page = input.Page
	}
	if input.PageSize > 0 {
		q.PageSize = input.PageSize
	}
	q.Search = input.Search
	q.SortBy = []string{"sort_order"}
	q.SortOrder = []string{"asc"}

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

func (h *TaxonomyHandler) HandleListTags(ctx context.Context, input *dto.ListTagsInput) (*dto.ListTagsOutput, error) {
	q := query.QueryOptions{}
	if input.Page > 0 {
		q.Page = input.Page
	}
	if input.PageSize > 0 {
		q.PageSize = input.PageSize
	}
	q.Search = input.Search
	q.Preload = []string{"Translations"}

	result := h.tagRepo.FindAll(ctx, q)
	items := make([]dto.TagResponse, 0, len(result.Data))
	for _, tag := range result.Data {
		items = append(items, tagToResponse(tag))
	}
	return &dto.ListTagsOutput{Body: dto.ListTagsResponseBody{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}}, nil
}

func (h *TaxonomyHandler) HandleGetSector(ctx context.Context, input *dto.GetSectorInput) (*dto.GetSectorOutput, error) {
	var sector entity.Sector
	if err := h.sectorRepo.GetDB().WithContext(ctx).
		Preload("Translations").First(&sector, "id = ?", input.ID).Error; err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetSectorOutput{Body: sectorToResponse(&sector)}, nil
}

func (h *TaxonomyHandler) HandleGetTag(ctx context.Context, input *dto.GetTagInput) (*dto.GetTagOutput, error) {
	var tag entity.Tag
	if err := h.tagRepo.GetDB().WithContext(ctx).
		Preload("Translations").First(&tag, "id = ?", input.ID).Error; err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetTagOutput{Body: tagToResponse(&tag)}, nil
}
