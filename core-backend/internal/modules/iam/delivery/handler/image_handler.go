package handler

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
)

type ImageHandler struct {
	avatarService *service.AvatarService
}

func NewImageHandler(avatarService *service.AvatarService) *ImageHandler {
	return &ImageHandler{
		avatarService: avatarService,
	}
}

func (h *ImageHandler) HandleUploadAvatar(ctx context.Context, input *dto.UploadAvatarInput) (*dto.UploadAvatarOutput, error) {
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
	if userID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.unauthorized"))
	}

	file, filename, err := h.extractFileFromContext(ctx)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("iam.errors.invalidFile"))
	}
	defer func() { _ = file.Close() }()

	fileBytes, err := h.readBoundedFile(file)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	result, err := h.avatarService.UploadAvatar(ctx, service.UploadAvatarInput{
		UserID:    userID,
		FileBytes: fileBytes,
		Filename:  filename,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UploadAvatarOutput{
		Body: dto.UploadAvatarResponse{
			ImageURL: result.ImageURL,
		},
	}, nil
}

func (h *ImageHandler) extractFileFromContext(ctx context.Context) (multipart.File, string, error) {
	hc, ok := ctx.Value("huma-context").(huma.Context)
	if !ok {
		if hc, ok = ctx.(huma.Context); !ok {
			return nil, "", fmt.Errorf("invalid context type")
		}
	}

	form, err := hc.GetMultipartForm()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get multipart form: %w", err)
	}

	if form.File == nil {
		return nil, "", fmt.Errorf("no file provided")
	}

	fileHeader, exists := form.File["file"]
	if !exists || len(fileHeader) == 0 {
		return nil, "", fmt.Errorf("no file provided")
	}

	file, err := fileHeader[0].Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return file, fileHeader[0].Filename, nil
}

func (h *ImageHandler) readBoundedFile(file multipart.File) ([]byte, error) {
	buffer := make([]byte, service.MaxAvatarSize+1)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return nil, apperrors.InternalError("iam.errors.readFileFailed", err)
	}

	if n > service.MaxAvatarSize {
		return nil, apperrors.PayloadTooLargeError("iam.errors.fileTooLarge")
	}

	return buffer[:n], nil
}
