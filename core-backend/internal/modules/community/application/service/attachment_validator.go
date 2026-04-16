package service

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	_ "golang.org/x/image/webp"
)

const MaxCommunityAttachmentSize = 10 << 20 // 10MB

var allowedCommunityAttachmentTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
}

type ValidatedCommunityAttachment struct {
	Content     []byte
	ContentType string
	Extension   string
}

type CommunityAttachmentValidator struct{}

func NewCommunityAttachmentValidator() *CommunityAttachmentValidator {
	return &CommunityAttachmentValidator{}
}

func (v *CommunityAttachmentValidator) Validate(fileBytes []byte) (*ValidatedCommunityAttachment, error) {
	return v.ValidateWithLimit(fileBytes, MaxCommunityAttachmentSize)
}

func (v *CommunityAttachmentValidator) ValidateWithLimit(fileBytes []byte, maxSize int64) (*ValidatedCommunityAttachment, error) {
	if len(fileBytes) == 0 {
		return nil, apperrors.BadRequestError("community.errors.invalidFile")
	}

	if int64(len(fileBytes)) > maxSize {
		return nil, apperrors.PayloadTooLargeError("community.errors.fileTooLarge")
	}

	sniffLen := len(fileBytes)
	if sniffLen > 512 {
		sniffLen = 512
	}

	contentType := strings.ToLower(http.DetectContentType(fileBytes[:sniffLen]))
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	ext, allowed := allowedCommunityAttachmentTypes[contentType]
	if !allowed {
		return nil, apperrors.BadRequestError("community.errors.invalidFileType")
	}

	if strings.HasPrefix(contentType, "image/") {
		if _, _, err := image.DecodeConfig(bytes.NewReader(fileBytes)); err != nil {
			return nil, apperrors.BadRequestError("community.errors.invalidFileType")
		}
	}

	if contentType == "application/pdf" && !bytes.HasPrefix(fileBytes, []byte("%PDF-")) {
		return nil, apperrors.BadRequestError("community.errors.invalidFileType")
	}

	return &ValidatedCommunityAttachment{
		Content:     fileBytes,
		ContentType: contentType,
		Extension:   ext,
	}, nil
}
