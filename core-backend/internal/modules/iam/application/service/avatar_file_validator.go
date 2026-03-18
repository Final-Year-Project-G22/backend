package service

import (
	"bytes"
	"image"
	"net/http"

	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

const MaxAvatarSize = 5 << 20 // 5MB

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type ValidatedAvatar struct {
	Content     []byte
	ContentType string
	Extension   string
}

type AvatarValidator struct{}

func NewAvatarValidator() *AvatarValidator {
	return &AvatarValidator{}
}

func (v *AvatarValidator) Validate(fileBytes []byte) (*ValidatedAvatar, error) {
	if len(fileBytes) > MaxAvatarSize {
		return nil, apperrors.PayloadTooLargeError("iam.errors.fileTooLarge")
	}

	contentType := http.DetectContentType(fileBytes)
	ext, allowed := allowedAvatarTypes[contentType]
	if !allowed {
		return nil, apperrors.BadRequestError("iam.errors.invalidFileType")
	}

	_, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, apperrors.BadRequestError("iam.errors.invalidFileType")
	}

	return &ValidatedAvatar{
		Content:     fileBytes,
		ContentType: contentType,
		Extension:   ext,
	}, nil
}

func (v *AvatarValidator) ValidateWithLimit(fileBytes []byte, maxSize int64) (*ValidatedAvatar, error) {
	if int64(len(fileBytes)) > maxSize {
		return nil, apperrors.PayloadTooLargeError("iam.errors.fileTooLarge")
	}

	contentType := http.DetectContentType(fileBytes)
	ext, allowed := allowedAvatarTypes[contentType]
	if !allowed {
		return nil, apperrors.BadRequestError("iam.errors.invalidFileType")
	}

	_, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, apperrors.BadRequestError("iam.errors.invalidFileType")
	}

	return &ValidatedAvatar{
		Content:     fileBytes,
		ContentType: contentType,
		Extension:   ext,
	}, nil
}
