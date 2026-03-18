package service

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	_ "golang.org/x/image/webp"

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
	return v.ValidateWithLimit(fileBytes, MaxAvatarSize)
}

func (v *AvatarValidator) ValidateWithLimit(fileBytes []byte, maxSize int64) (*ValidatedAvatar, error) {
	if len(fileBytes) == 0 {
		return nil, apperrors.BadRequestError("iam.errors.invalidFile")
	}

	if int64(len(fileBytes)) > maxSize {
		return nil, apperrors.PayloadTooLargeError("iam.errors.fileTooLarge")
	}

	sniffLen := len(fileBytes)
	if sniffLen > 512 {
		sniffLen = 512
	}

	contentType := strings.ToLower(http.DetectContentType(fileBytes[:sniffLen]))
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

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
