package handler

import (
	"context"
	"io"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type BusinessProfileImageHandler struct {
	bpImageService *service.BusinessProfileImageService
}

func NewBusinessProfileImageHandler(bpImageService *service.BusinessProfileImageService) *BusinessProfileImageHandler {
	return &BusinessProfileImageHandler{
		bpImageService: bpImageService,
	}
}

func (h *BusinessProfileImageHandler) HandleUploadBusinessLogo(ctx context.Context, input *dto.UploadBusinessImageInput) (*dto.UploadBusinessImageOutput, error) {
	return h.handleUpload(ctx, input, "logo")
}

func (h *BusinessProfileImageHandler) HandleUploadBusinessBanner(ctx context.Context, input *dto.UploadBusinessImageInput) (*dto.UploadBusinessImageOutput, error) {
	return h.handleUpload(ctx, input, "banner")
}

func (h *BusinessProfileImageHandler) handleUpload(ctx context.Context, input *dto.UploadBusinessImageInput, imageType string) (*dto.UploadBusinessImageOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
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

	limitedReader := io.LimitReader(file, int64(service.MaxBusinessImageSize)+1)
	fileBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.InternalError("iam.errors.readFileFailed", err))
	}

	if len(fileBytes) > service.MaxBusinessImageSize {
		return nil, apperrors.ToHumaError(ctx, apperrors.PayloadTooLargeError("iam.errors.fileTooLarge"))
	}

	result, err := h.bpImageService.UploadBusinessImage(ctx, service.UploadBusinessImageInput{
		AccountID: accountID,
		FileBytes: fileBytes,
		Filename:  formData.File.Filename,
		ImageType: imageType,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UploadBusinessImageOutput{
		Body: dto.UploadBusinessImageResponse{
			ImageURL: result.ImageURL,
		},
	}, nil
}
