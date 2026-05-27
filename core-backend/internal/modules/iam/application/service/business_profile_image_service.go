package service

import (
	"context"
	"fmt"

	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
)

const MaxBusinessImageSize = 5 << 20 // 5MB

type UploadBusinessImageInput struct {
	AccountID uuid.UUID
	FileBytes []byte
	Filename  string
	ImageType string // "logo" or "banner"
}

type UploadBusinessImageOutput struct {
	ImageURL string
}

type BusinessProfileImageService struct {
	storage         storage.Storage
	avatarValidator *AvatarValidator
}

func NewBusinessProfileImageService(
	storage storage.Storage,
	avatarValidator *AvatarValidator,
) *BusinessProfileImageService {
	return &BusinessProfileImageService{
		storage:         storage,
		avatarValidator: avatarValidator,
	}
}

func (s *BusinessProfileImageService) UploadBusinessImage(ctx context.Context, input UploadBusinessImageInput) (*UploadBusinessImageOutput, error) {
	validated, err := s.avatarValidator.ValidateWithLimit(input.FileBytes, MaxBusinessImageSize)
	if err != nil {
		return nil, err
	}

	keyPrefix := "logos"
	if input.ImageType == "banner" {
		keyPrefix = "banners"
	}
	key := fmt.Sprintf("business/%s/%s/%s%s", keyPrefix, input.AccountID.String(), uuid.New().String(), validated.Extension)

	uploaded, err := s.storage.Upload(ctx, storage.UploadOptions{
		Key:         key,
		Content:     validated.Content,
		ContentType: validated.ContentType,
	})
	if err != nil {
		return nil, apperrors.InternalError("iam.errors.uploadFailed", fmt.Errorf("storage upload key=%s contentType=%s size=%d: %w", key, validated.ContentType, len(validated.Content), err))
	}

	imageURL := uploaded.URL
	if imageURL == "" {
		imageURL = fmt.Sprintf("/api/v1/files/%s", key)
	}

	return &UploadBusinessImageOutput{
		ImageURL: imageURL,
	}, nil
}
