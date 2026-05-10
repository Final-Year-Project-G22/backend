package handler

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type IngestionHandler struct {
	ingestionService *service.IngestionService
}

func NewIngestionHandler(ingestionService *service.IngestionService) *IngestionHandler {
	return &IngestionHandler{ingestionService: ingestionService}
}

func (h *IngestionHandler) Healthcheck(ctx context.Context) error {
	return h.ingestionService.Ping(ctx)
}

func (h *IngestionHandler) HandleCreateUploadIntent(ctx context.Context, input *dto.CreateUploadIntentInput) (*dto.CreateUploadIntentOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ingestion.errors.unauthorized"))
	}

	var expiresIn *time.Duration
	if input.Body.ExpiresInSec != nil {
		d := time.Duration(*input.Body.ExpiresInSec) * time.Second
		expiresIn = &d
	}

	out, err := h.ingestionService.CreateUploadIntent(ctx, service.CreateUploadIntentInput{
		StorageKey:  input.Body.StorageKey,
		ContentType: input.Body.ContentType,
		Metadata:    input.Body.Metadata,
		ExpiresIn:   expiresIn,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.CreateUploadIntentOutput{
		Body: dto.CreateUploadIntentResponseBody{
			Key:       out.Key,
			UploadURL: out.UploadURL,
			Method:    out.Method,
			Headers:   out.Headers,
			ExpiresAt: out.ExpiresAt,
		},
	}, nil
}

func (h *IngestionHandler) HandleFinalizeUpload(ctx context.Context, input *dto.FinalizeUploadInput) (*dto.FinalizeUploadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ingestion.errors.unauthorized"))
	}
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
	if userID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ingestion.errors.unauthorized"))
	}

	out, err := h.ingestionService.FinalizeUpload(ctx, service.FinalizeUploadInput{
		AccountID:        accountID,
		UserID:           userID,
		StorageKey:       input.Body.StorageKey,
		ContentType:      input.Body.ContentType,
		SizeBytes:        input.Body.SizeBytes,
		ChecksumSHA256:   input.Body.ChecksumSHA256,
		IdempotencyKey:   input.Body.IdempotencyKey,
		SourceFilename:   input.Body.SourceFilename,
		DeclaredLanguage: input.Body.DeclaredLanguage,
		BatchID:          input.Body.BatchID,
		SectorIDs:        input.Body.SectorIDs,
		TagIDs:           input.Body.TagIDs,
		Region:           input.Body.Region,
		Stage:            input.Body.Stage,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.FinalizeUploadOutput{
		Body: dto.FinalizeUploadResponseBody{
			IngestionID: out.IngestionID,
			DocumentID:  out.DocumentID,
			EventID:     out.EventID,
			State:       out.State,
		},
	}, nil
}

func (h *IngestionHandler) HandleDeleteDocument(ctx context.Context, input *dto.DeleteDocumentInput) (*dto.DeleteDocumentOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ingestion.errors.unauthorized"))
	}

	if err := h.ingestionService.DeleteDocument(ctx, input.DocumentID, accountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.DeleteDocumentOutput{
		Body: dto.DeleteDocumentResponseBody{Success: true},
	}, nil
}
