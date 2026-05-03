package coregrpc

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	coredocumentv1 "github.com/Final-Year-Project-G22/backend/core/pb/core/document/v1"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DocumentFetchService implements the gRPC DocumentFetchService.
type DocumentFetchService struct {
	coredocumentv1.UnimplementedDocumentFetchServiceServer
	storage storage.Storage
	docRepo airepository.IngestionDocumentRepository
	logger  core.Logger
}

// NewDocumentFetchService creates a new DocumentFetchService.
func NewDocumentFetchService(
	storage storage.Storage,
	docRepo airepository.IngestionDocumentRepository,
	logger core.Logger,
) *DocumentFetchService {
	return &DocumentFetchService{
		storage: storage,
		docRepo: docRepo,
		logger:  logger,
	}
}

// GetSignedUrl generates a presigned URL for downloading an ingestion document.
func (s *DocumentFetchService) GetSignedUrl(ctx context.Context, req *coredocumentv1.GetSignedUrlRequest) (*coredocumentv1.GetSignedUrlResponse, error) {
	documentID, err := uuid.Parse(req.GetDocumentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid document_id")
	}

	accountID, err := uuid.Parse(req.GetAccountId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}

	doc, err := s.docRepo.GetByID(ctx, documentID)
	if err != nil {
		s.logger.Error("failed to get ingestion document", core.Error(err), core.String("document_id", documentID.String()))
		return nil, status.Error(codes.Internal, "failed to get document")
	}
	if doc == nil {
		return nil, status.Error(codes.NotFound, "document not found")
	}

	// Validate ownership
	if doc.AccountID != accountID {
		s.logger.Warn("unauthorized document access attempt",
			core.String("document_id", documentID.String()),
			core.String("requested_account", accountID.String()),
			core.String("actual_account", doc.AccountID.String()),
		)
		return nil, status.Error(codes.PermissionDenied, "document does not belong to account")
	}

	// Generate presigned URL
	expiry := time.Duration(req.GetExpiresInSeconds()) * time.Second
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	signedURL, err := s.storage.GetPresignedURL(ctx, doc.StorageKey, expiry)
	if err != nil {
		s.logger.Error("failed to generate presigned URL", core.Error(err), core.String("storage_key", doc.StorageKey))
		return nil, status.Error(codes.Internal, "failed to generate download URL")
	}

	return &coredocumentv1.GetSignedUrlResponse{
		SignedUrl:     signedURL,
		ExpiresAt:     time.Now().Add(expiry).Unix(),
		ContentType:   doc.ContentType,
		ContentLength: doc.SizeBytes,
	}, nil
}
