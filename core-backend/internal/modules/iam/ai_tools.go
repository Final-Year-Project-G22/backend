package iam

import (
	"context"
	"encoding/json"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ListSectorsTool struct {
	sectorRepo repository.SectorRepository
}

func NewListSectorsTool(sectorRepo repository.SectorRepository) *ListSectorsTool {
	return &ListSectorsTool{sectorRepo: sectorRepo}
}

func (t *ListSectorsTool) Name() string { return "list_sectors" }

func (t *ListSectorsTool) Description() string {
	return "List all business sectors with their translations."
}

func (t *ListSectorsTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {}
	}`
}

func (t *ListSectorsTool) Execute(ctx context.Context, _ string, _, _ uuid.UUID) (string, error) {
	opts := query.DefaultQueryOptions()
	opts.PageSize = 200
	opts.Preload = []string{"Translations"}

	result := t.sectorRepo.FindAll(ctx, opts)
	payload, _ := json.Marshal(result.Data)
	return string(payload), nil
}

type listTagsArgs struct {
	Group *string `json:"group,omitempty"`
}

type ListTagsTool struct {
	tagRepo repository.TagRepository
}

func NewListTagsTool(tagRepo repository.TagRepository) *ListTagsTool {
	return &ListTagsTool{tagRepo: tagRepo}
}

func (t *ListTagsTool) Name() string { return "list_tags" }

func (t *ListTagsTool) Description() string {
	return "List all tags, optionally filtered by group (e.g., business_stage)."
}

func (t *ListTagsTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {
			"group": {"type": "string", "description": "Filter tags by group name (e.g., business_stage)"}
		}
	}`
}

func (t *ListTagsTool) Execute(ctx context.Context, argsJSON string, _, _ uuid.UUID) (string, error) {
	var args listTagsArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	opts := query.DefaultQueryOptions()
	opts.PageSize = 200
	opts.Preload = []string{"Translations"}
	if args.Group != nil && *args.Group != "" {
		opts.Filters = map[string]interface{}{"group": *args.Group}
	}

	result := t.tagRepo.FindAll(ctx, opts)
	payload, _ := json.Marshal(result.Data)
	return string(payload), nil
}

// GetUserProfileTool returns the user's business profile including sector, region, stage, and tags.
type GetUserProfileTool struct {
	bpRepo repository.BusinessProfileRepository
}

func NewGetUserProfileTool(bpRepo repository.BusinessProfileRepository) *GetUserProfileTool {
	return &GetUserProfileTool{bpRepo: bpRepo}
}

func (t *GetUserProfileTool) Name() string { return "get_user_profile" }

func (t *GetUserProfileTool) Description() string {
	return "Get the user's business profile: company name, sector, region, business stage, and tags."
}

func (t *GetUserProfileTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {}
	}`
}

func (t *GetUserProfileTool) Execute(ctx context.Context, _ string, accountID, _ uuid.UUID) (string, error) {
	profile, err := t.bpRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return "", err
	}
	if profile == nil {
		return `{"error": "no business profile found"}`, nil
	}

	type profileResult struct {
		CompanyName string `json:"companyName"`
		Region      string `json:"region,omitempty"`
		Stage       string `json:"stage,omitempty"`
		SectorID    string `json:"sectorId,omitempty"`
		TagCount    int    `json:"tagCount"`
	}
	result := profileResult{
		CompanyName: profile.CompanyName,
		TagCount:    len(profile.Tags),
	}
	if profile.Region != nil {
		result.Region = string(*profile.Region)
	}
	if profile.Stage != nil {
		result.Stage = string(*profile.Stage)
	}
	if profile.SectorID != nil {
		result.SectorID = profile.SectorID.String()
	}

	payload, _ := json.Marshal(result)
	return string(payload), nil
}
