package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadAvatarInput struct {
	UserID    uuid.UUID
	FileBytes []byte
	Filename  string
}

type UploadAvatarOutput struct {
	ImageURL string
}

type AvatarService struct {
	userRepo        repository.UserRepository
	storage         storage.Storage
	avatarValidator *AvatarValidator
}

func NewAvatarService(
	userRepo repository.UserRepository,
	storage storage.Storage,
	avatarValidator *AvatarValidator,
) *AvatarService {
	return &AvatarService{
		userRepo:        userRepo,
		storage:         storage,
		avatarValidator: avatarValidator,
	}
}

func (s *AvatarService) UploadAvatar(ctx context.Context, input UploadAvatarInput) (*UploadAvatarOutput, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFoundError("user", input.UserID)
		}
		return nil, apperrors.InternalError("iam.errors.userNotFound", err)
	}
	_ = user

	oldImageURL, err := s.userRepo.GetImageURL(input.UserID.String())
	if err != nil {
		return nil, err
	}

	validated, err := s.avatarValidator.Validate(input.FileBytes)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("avatars/%s%s", input.UserID.String(), validated.Extension)

	uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
		Key:         key,
		Content:     validated.Content,
		ContentType: validated.ContentType,
	})
	if err != nil {
		return nil, apperrors.InternalError("iam.errors.uploadFailed", err)
	}

	imageURL := uploaded.URL
	if imageURL == "" {
		imageURL = fmt.Sprintf("/api/v1/files/%s", key)
	}

	if err := s.userRepo.UpdateAvatar(input.UserID.String(), imageURL); err != nil {
		_ = s.storage.Delete(ctx, key)
		return nil, apperrors.InternalError("iam.errors.updateFailed", err)
	}

	if oldImageURL != "" {
		oldKey := s.extractKeyFromURL(oldImageURL)
		if oldKey != "" && oldKey != key {
			_ = s.storage.Delete(ctx, oldKey)
		}
	}

	return &UploadAvatarOutput{
		ImageURL: imageURL,
	}, nil
}

func (s *AvatarService) extractKeyFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return ""
}
