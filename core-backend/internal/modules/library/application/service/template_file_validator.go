package service

import (
	"bytes"
	"path/filepath"
	"strings"

	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

const maxFileSize int64 = 10 * 1024 * 1024

type FileValidationResult struct {
	Content     []byte
	ContentType string
	Extension   string
}

type TemplateFileValidator struct{}

func NewTemplateFileValidator() *TemplateFileValidator {
	return &TemplateFileValidator{}
}

type fileTypeInfo struct {
	magic     []byte
	mimeType  string
	extension string
}

var allowedFileTypes = map[string]fileTypeInfo{
	"pdf":  {magic: []byte("%PDF-"), mimeType: "application/pdf", extension: ".pdf"},
	"docx": {magic: []byte("PK\x03\x04"), mimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", extension: ".docx"},
	"xlsx": {magic: []byte("PK\x03\x04"), mimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", extension: ".xlsx"},
}

func (v *TemplateFileValidator) Validate(content []byte, filename string) (*FileValidationResult, error) {
	if len(content) == 0 {
		return nil, apperrors.InvalidInputError("file", "library.errors.invalidFileType")
	}

	if int64(len(content)) > maxFileSize {
		return nil, apperrors.PayloadTooLargeError("library.errors.fileTooLarge")
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	info, ok := allowedFileTypes[ext]
	if !ok {
		return nil, apperrors.InvalidInputError("file", "library.errors.invalidFileType")
	}

	if !bytes.HasPrefix(content, info.magic) {
		return nil, apperrors.InvalidInputError("file", "library.errors.invalidFileType")
	}

	return &FileValidationResult{
		Content:     content,
		ContentType: info.mimeType,
		Extension:   info.extension,
	}, nil
}

func (v *TemplateFileValidator) ValidateThumbnail(content []byte, filename string) error {
	if int64(len(content)) > 2*1024*1024 {
		return apperrors.PayloadTooLargeError("library.errors.fileTooLarge")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return nil
	default:
		return apperrors.InvalidInputError("thumbnail", "library.errors.invalidFileType")
	}
}
