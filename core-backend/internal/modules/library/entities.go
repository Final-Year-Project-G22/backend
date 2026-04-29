package library

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
)

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (e *EntityProvider) Entities() []any {
	return []any{
		&entity.LibraryCategory{},
		&entity.LibraryCategoryTranslation{},
		&entity.LibraryTemplateGroup{},
		&entity.LibraryTemplate{},
		&entity.LibraryInteractiveForm{},
		&entity.LibraryTemplateDownload{},
	}
}

func (e *EntityProvider) ModuleName() string {
	return "library"
}
