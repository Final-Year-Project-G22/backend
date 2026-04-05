package guide

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
)

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (e *EntityProvider) Entities() []any {
	return []any{
		&entity.GuideCategory{},
		&entity.Guide{},
		&entity.GuideCondition{},
		&entity.GuideCategoryCondition{},
		&entity.GuideStep{},
		&entity.StepCondition{},
		&entity.StepDependency{},
		&entity.UserGuideProgress{},
		&entity.UserGuideJourney{},
		&entity.UserGuideBookmark{},
		&entity.GuideStepVersion{},
		&entity.GuideCategoryTranslation{},
		&entity.GuideTranslation{},
		&entity.GuideStepTranslation{},
		&entity.UserGuideRecentView{},
	}
}

func (e *EntityProvider) ModuleName() string {
	return "guide"
}
