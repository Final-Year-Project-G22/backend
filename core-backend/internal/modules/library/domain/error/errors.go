package error

import "errors"

var (
	ErrTemplateNotFound        = errors.New("library: template not found")
	ErrTemplateGroupNotFound   = errors.New("library: template group not found")
	ErrCategoryNotFound        = errors.New("library: category not found")
	ErrInvalidFileType         = errors.New("library: invalid file type")
	ErrFileTooLarge            = errors.New("library: file too large")
	ErrTierAccessDenied        = errors.New("library: tier access denied")
	ErrAuthRequired            = errors.New("library: authentication required")
	ErrSlugAlreadyExists       = errors.New("library: slug already exists")
	ErrCategoryHasActiveGroups = errors.New("library: category has active template groups")
)
