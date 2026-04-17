package ai

import "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (p *EntityProvider) Entities() []any {
	return []any{
		&entity.IngestionDocument{},
		&entity.IngestionOutbox{},
		&entity.IngestionStatusEvent{},
		&entity.IngestionStatusProjection{},
	}
}

func (p *EntityProvider) ModuleName() string {
	return "ai"
}
