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
