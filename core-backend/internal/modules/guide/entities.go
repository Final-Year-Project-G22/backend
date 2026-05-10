package guide

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
)

func init() {
	query.RegisterConfig("Guide", query.EntityConfig{
		SearchableColumns: []string{"slug"},
		SortableColumns:   []string{"slug", "sort_order", "created_at"},
		DefaultSort:       []string{"sort_order", "created_at"},
	})

	query.RegisterConfig("GuideStep", query.EntityConfig{
		SearchableColumns: []string{"slug"},
		SortableColumns:   []string{"slug", "sort_order", "created_at"},
		DefaultSort:       []string{"sort_order", "created_at"},
	})
}

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (e *EntityProvider) Entities() []any {
	return []any{
		&entity.Guide{},
		&entity.GuideCondition{},
		&entity.GuideStep{},
		&entity.StepCondition{},
		&entity.StepDependency{},
		&entity.UserGuideProgress{},
		&entity.UserGuideJourney{},
		&entity.UserGuideBookmark{},
		&entity.GuideStepVersion{},
		&entity.GuideTranslation{},
		&entity.GuideStepTranslation{},
		&entity.UserGuideRecentView{},
	}
}

func (e *EntityProvider) ModuleName() string {
	return "guide"
}
