package guide

import (
	"context"
	"encoding/json"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type searchGuidesArgs struct {
	Keyword   *string     `json:"keyword,omitempty"`
	SectorIDs []uuid.UUID `json:"sectorIds,omitempty"`
	TagIDs    []uuid.UUID `json:"tagIds,omitempty"`
}

type SearchGuidesTool struct {
	guideViewUC usecase.GuideViewUseCase
}

func NewSearchGuidesTool(guideViewUC usecase.GuideViewUseCase) *SearchGuidesTool {
	return &SearchGuidesTool{guideViewUC: guideViewUC}
}

func (t *SearchGuidesTool) Name() string { return "search_guides" }

func (t *SearchGuidesTool) Description() string {
	return "Search business formalization guides by keyword, sectors, or tags."
}

func (t *SearchGuidesTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {
			"keyword": {"type": "string", "description": "Search keyword"},
			"sectorIds": {"type": "array", "items": {"type": "string", "format": "uuid"}, "description": "Filter by sector IDs"},
			"tagIds": {"type": "array", "items": {"type": "string", "format": "uuid"}, "description": "Filter by tag IDs"}
		}
	}`
}

func (t *SearchGuidesTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
	var args searchGuidesArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	opts := query.DefaultQueryOptions()
	opts.Page = 1
	opts.PageSize = 50
	opts.Filters = make(map[string]interface{})
	if len(args.SectorIDs) > 0 {
		opts.Filters["sector_ids"] = args.SectorIDs
	}
	if len(args.TagIDs) > 0 {
		opts.Filters["tag_ids"] = args.TagIDs
	}

	var guides []*usecase.GuideCard
	var err error
	if args.Keyword != nil && *args.Keyword != "" {
		guides, err = t.guideViewUC.SearchGuides(ctx, accountID, userID, *args.Keyword, opts, constants.LocaleEnglish)
	} else {
		guides, err = t.guideViewUC.ListGuides(ctx, accountID, userID, opts, constants.LocaleEnglish, args.SectorIDs, args.TagIDs)
	}
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(guides)
	return string(result), nil
}

type guideDetailArgs struct {
	GuideID string `json:"guideId"`
}

type GuideDetailTool struct {
	guideAdminUC usecase.GuideManagementUseCase
}

func NewGuideDetailTool(guideAdminUC usecase.GuideManagementUseCase) *GuideDetailTool {
	return &GuideDetailTool{guideAdminUC: guideAdminUC}
}

func (t *GuideDetailTool) Name() string { return "get_guide_detail" }

func (t *GuideDetailTool) Description() string {
	return "Get full details of a business formalization guide including all steps."
}

func (t *GuideDetailTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {
			"guideId": {"type": "string", "format": "uuid", "description": "Guide ID"}
		},
		"required": ["guideId"]
	}`
}

func (t *GuideDetailTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
	var args guideDetailArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	guideID, err := uuid.Parse(args.GuideID)
	if err != nil {
		return "", err
	}

	guide, err := t.guideAdminUC.GetGuideDetail(ctx, guideID, constants.LocaleEnglish)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(guide)
	return string(result), nil
}
