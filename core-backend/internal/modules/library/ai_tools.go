package library

import (
	"context"
	"encoding/json"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type findTemplateArgs struct {
	Query  string  `json:"query"`
	Source *string `json:"source,omitempty"`
}

type FindTemplateTool struct {
	templateRepo repository.LibraryTemplateRepository
	groupRepo    repository.LibraryTemplateGroupRepository
}

func NewFindTemplateTool(
	templateRepo repository.LibraryTemplateRepository,
	groupRepo repository.LibraryTemplateGroupRepository,
) *FindTemplateTool {
	return &FindTemplateTool{templateRepo: templateRepo, groupRepo: groupRepo}
}

func (t *FindTemplateTool) Name() string { return "find_template" }

func (t *FindTemplateTool) Description() string {
	return "Find document templates such as business registration forms, tax filing forms, and license applications available in the library."
}

func (t *FindTemplateTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {
			"query": {"type": "string", "description": "Search keyword for template name or description"},
			"source": {"type": "string", "description": "Optional template source filter"}
		},
		"required": ["query"]
	}`
}

func (t *FindTemplateTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
	var args findTemplateArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	opts := query.DefaultQueryOptions()
	opts.PageSize = 20

	templates := t.templateRepo.FindAll(ctx, opts)
	payload, _ := json.Marshal(templates.Data)
	return string(payload), nil
}
