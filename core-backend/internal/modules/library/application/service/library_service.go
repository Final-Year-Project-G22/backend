package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
)

type LibraryService interface {
	CreateTemplateGroup(ctx context.Context, accountID uuid.UUID, input usecase.CreateTemplateGroupInput) (*entity.LibraryTemplateGroup, error)
	UpdateTemplateGroup(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateGroupInput) (*entity.LibraryTemplateGroup, error)
	CreateTemplate(ctx context.Context, input usecase.CreateTemplateInput) (*entity.LibraryTemplate, error)
	UpdateTemplate(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateInput) (*entity.LibraryTemplate, error)
}

type libraryService struct {
	adminUC   usecase.LibraryAdminUsecase
	storage   storage.Storage
	validator *TemplateFileValidator
}

func NewLibraryService(
	adminUC usecase.LibraryAdminUsecase,
	storage storage.Storage,
	validator *TemplateFileValidator,
) LibraryService {
	return &libraryService{
		adminUC:   adminUC,
		storage:   storage,
		validator: validator,
	}
}

func (s *libraryService) CreateTemplateGroup(ctx context.Context, accountID uuid.UUID, input usecase.CreateTemplateGroupInput) (*entity.LibraryTemplateGroup, error) {
	group, err := s.adminUC.CreateTemplateGroup(ctx, accountID, input)
	if err != nil {
		return nil, err
	}

	if input.ThumbnailBytes != nil && input.ThumbnailFilename != nil {
		if err := s.validator.ValidateThumbnail(input.ThumbnailBytes, *input.ThumbnailFilename); err != nil {
			return nil, err
		}
		ext := strings.ToLower(filepath.Ext(*input.ThumbnailFilename))
		key := fmt.Sprintf("library/thumbnails/%s%s", group.ID.String(), ext)
		uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
			Key:     key,
			Content: input.ThumbnailBytes,
		})
		if err != nil {
			return nil, apperrors.InternalError("library.errors.uploadFailed", err)
		}
		url := uploaded.URL
		if url == "" {
			url = fmt.Sprintf("/api/v1/files/%s", key)
		}
		group.ThumbnailURL = &url
		if _, err := s.adminUC.UpdateTemplateGroup(ctx, group.ID, usecase.UpdateTemplateGroupInput{
			ThumbnailURL: &url,
		}); err != nil {
			_ = s.storage.Delete(ctx, key)
			return nil, err
		}
	}

	return group, nil
}

func (s *libraryService) UpdateTemplateGroup(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateGroupInput) (*entity.LibraryTemplateGroup, error) {
	if input.ThumbnailBytes != nil && input.ThumbnailFilename != nil {
		if err := s.validator.ValidateThumbnail(input.ThumbnailBytes, *input.ThumbnailFilename); err != nil {
			return nil, err
		}
		current, err := s.adminUC.GetTemplateGroup(ctx, id)
		if err != nil {
			return nil, err
		}
		if current.ThumbnailURL != nil {
			oldKey := s.extractKeyFromURL(*current.ThumbnailURL)
			if oldKey != "" {
				_ = s.storage.Delete(ctx, oldKey)
			}
		}

		ext := strings.ToLower(filepath.Ext(*input.ThumbnailFilename))
		key := fmt.Sprintf("library/thumbnails/%s%s", id.String(), ext)
		uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
			Key:     key,
			Content: input.ThumbnailBytes,
		})
		if err != nil {
			return nil, apperrors.InternalError("library.errors.uploadFailed", err)
		}
		url := uploaded.URL
		if url == "" {
			url = fmt.Sprintf("/api/v1/files/%s", key)
		}
		input.ThumbnailURL = &url
	}

	return s.adminUC.UpdateTemplateGroup(ctx, id, input)
}

func (s *libraryService) CreateTemplate(ctx context.Context, input usecase.CreateTemplateInput) (*entity.LibraryTemplate, error) {
	validated, err := s.validator.Validate(input.FileBytes, input.Filename)
	if err != nil {
		return nil, err
	}

	fileKey := fmt.Sprintf("library/templates/%s%s", uuid.New().String(), validated.Extension)
	uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
		Key:         fileKey,
		Content:     validated.Content,
		ContentType: validated.ContentType,
	})
	if err != nil {
		return nil, apperrors.InternalError("library.errors.uploadFailed", err)
	}

	url := uploaded.URL
	if url == "" {
		url = fmt.Sprintf("/api/v1/files/%s", fileKey)
	}

	tmpl, err := s.adminUC.CreateTemplate(ctx, usecase.CreateTemplateInput{
		GroupID:     input.GroupID,
		Language:    input.Language,
		Title:       input.Title,
		Description: input.Description,
		FileBytes:   input.FileBytes,
		Filename:    input.Filename,
		FileKey:     fileKey,
		FileURL:     &url,
		FileSize:    int64(len(input.FileBytes)),
		ContentType: validated.ContentType,
	})
	if err != nil {
		_ = s.storage.Delete(ctx, fileKey)
		return nil, err
	}

	return tmpl, nil
}

func (s *libraryService) UpdateTemplate(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateInput) (*entity.LibraryTemplate, error) {
	if input.FileBytes != nil && input.Filename != nil {
		validated, err := s.validator.Validate(input.FileBytes, *input.Filename)
		if err != nil {
			return nil, err
		}

		current, err := s.adminUC.GetTemplate(ctx, id)
		if err != nil {
			return nil, err
		}

		if current.FileKey != "" {
			_ = s.storage.Delete(ctx, current.FileKey)
		}

		fileKey := fmt.Sprintf("library/templates/%s%s", id.String(), validated.Extension)
		uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
			Key:         fileKey,
			Content:     validated.Content,
			ContentType: validated.ContentType,
		})
		if err != nil {
			return nil, apperrors.InternalError("library.errors.uploadFailed", err)
		}

		url := uploaded.URL
		if url == "" {
			url = fmt.Sprintf("/api/v1/files/%s", fileKey)
		}

		return s.adminUC.UpdateTemplate(ctx, id, usecase.UpdateTemplateInput{
			FileBytes:   input.FileBytes,
			Filename:    input.Filename,
			Title:       input.Title,
			Description: input.Description,
			IsActive:    input.IsActive,
			FileKey:     &fileKey,
			FileURL:     &url,
			FileSize:    int64Ptr(int64(len(input.FileBytes))),
			ContentType: &validated.ContentType,
		})
	}

	return s.adminUC.UpdateTemplate(ctx, id, input)
}

func int64Ptr(v int64) *int64 {
	return &v
}

func (s *libraryService) extractKeyFromURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.Contains(url, "/api/v1/files/") {
		parts := strings.Split(url, "/api/v1/files/")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}
	candidate := strings.Join(parts[len(parts)-2:], "/")
	if strings.HasPrefix(candidate, "library/") {
		return candidate
	}
	return ""
}
