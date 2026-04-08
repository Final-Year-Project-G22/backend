package community

import "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (e *EntityProvider) Entities() []any {
	return []any{
		&entity.CommunityCategory{},
		&entity.DiscussionThread{},
		&entity.DiscussionPost{},
		&entity.UserThreadSettings{},
		&entity.UserCategorySettings{},
		&entity.ContentReport{},
		&entity.ThreadBlockedUser{},
	}
}

func (e *EntityProvider) ModuleName() string {
	return "community"
}
