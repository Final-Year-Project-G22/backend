package error

import "errors"

var (
	ErrCategoryNotFound           = errors.New("guide: category not found")
	ErrGuideNotFound              = errors.New("guide: guide not found")
	ErrStepNotFound               = errors.New("guide: step not found")
	ErrCategoryConditionNotFound  = errors.New("guide: category condition not found")
	ErrGuideConditionNotFound     = errors.New("guide: guide condition not found")
	ErrStepConditionNotFound      = errors.New("guide: step condition not found")
	ErrDependencyNotFound         = errors.New("guide: dependency not found")
	ErrTranslationNotFound        = errors.New("guide: translation not found")
	ErrProgressNotFound           = errors.New("guide: progress not found")
	ErrJourneyNotFound            = errors.New("guide: journey not found")
	ErrBookmarkNotFound           = errors.New("guide: bookmark not found")
	ErrStepVersionNotFound        = errors.New("guide: step version not found")
	ErrVersionRestoreNotSupported = errors.New("guide: version restore not supported by current step model")
	ErrInvalidStatusTransition    = errors.New("guide: invalid status transition")
	ErrDependenciesNotMet         = errors.New("guide: dependencies not met")
	ErrStepNotOptional            = errors.New("guide: step not optional")
	ErrDependencyCycle            = errors.New("guide: dependency cycle detected")
	ErrIncompleteStepList         = errors.New("guide: incomplete step list for reorder")
)
