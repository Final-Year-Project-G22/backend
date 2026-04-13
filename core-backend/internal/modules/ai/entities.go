package ai

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (p *EntityProvider) Entities() []any {
	return []any{}
}

func (p *EntityProvider) ModuleName() string {
	return "ai"
}
