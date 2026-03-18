package handler

import (
	"context"
	"io"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
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

	formData := input.RawBody.Data()
	if formData == nil || !formData.File.IsSet {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("iam.errors.invalidFile"))
	}

	file := formData.File.File
	if file == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("iam.errors.invalidFile"))
	}
	defer func() { _ = file.Close() }()

	limitedReader := io.LimitReader(file, int64(service.MaxAvatarSize)+1)
	fileBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.InternalError("iam.errors.readFileFailed", err))
	}

	if len(fileBytes) > service.MaxAvatarSize {
		return nil, apperrors.ToHumaError(ctx, apperrors.PayloadTooLargeError("iam.errors.fileTooLarge"))
	}

	result, err := h.avatarService.UploadAvatar(ctx, service.UploadAvatarInput{
		UserID:    userID,
		FileBytes: fileBytes,
		Filename:  formData.File.Filename,
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
