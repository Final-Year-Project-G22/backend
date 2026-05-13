package service

import (
	"context"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	aievent "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/event"
	airepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	defaultUploadIntentTTL = 15 * time.Minute
	minUploadIntentTTL     = 60 * time.Second
	maxUploadIntentTTL     = 60 * time.Minute
)

type IngestionService struct {
	enabled        bool
	storage        storage.Storage
	documentRepo   airepo.IngestionDocumentRepository
	outboxRepo     airepo.IngestionOutboxRepository
	projectionRepo airepo.IngestionStatusProjectionRepository
	transactor     sharedrepo.Transactor
}

type CreateUploadIntentInput struct {
	StorageKey  *string
	ContentType string
	Metadata    map[string]string
	ExpiresIn   *time.Duration
}

type CreateUploadIntentOutput struct {
	Key       string
	UploadURL string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

type FinalizeUploadInput struct {
	AccountID        uuid.UUID
	UserID           uuid.UUID
	StorageKey       string
	ContentType      string
	SizeBytes        int64
	ChecksumSHA256   string
	IdempotencyKey   string
	SourceFilename   *string
	DeclaredLanguage *string
	BatchID          *uuid.UUID
	SectorIDs        []uuid.UUID
	TagIDs           []uuid.UUID
	Region           *string
	Stage            *string
}

type FinalizeUploadOutput struct {
	IngestionID uuid.UUID
	DocumentID  uuid.UUID
	EventID     uuid.UUID
	State       string
}

func NewIngestionService(
	enabled bool,
	storage storage.Storage,
	documentRepo airepo.IngestionDocumentRepository,
	outboxRepo airepo.IngestionOutboxRepository,
	projectionRepo airepo.IngestionStatusProjectionRepository,
	transactor sharedrepo.Transactor,
) *IngestionService {
	return &IngestionService{
		enabled:        enabled,
		storage:        storage,
		documentRepo:   documentRepo,
		outboxRepo:     outboxRepo,
		projectionRepo: projectionRepo,
		transactor:     transactor,
	}
}

func (s *IngestionService) Ping(_ context.Context) error {
	return nil
}

func (s *IngestionService) CreateUploadIntent(ctx context.Context, in CreateUploadIntentInput) (*CreateUploadIntentOutput, error) {
	if !s.enabled {
		return nil, apperrors.ForbiddenError("ingestion.errors.ingestionDisabled")
	}

	contentType := strings.TrimSpace(in.ContentType)
	if contentType == "" {
		return nil, apperrors.BadRequestError("ingestion.errors.contentTypeRequired")
	}

	ttl := defaultUploadIntentTTL
	if in.ExpiresIn != nil {
		ttl = *in.ExpiresIn
		if ttl < minUploadIntentTTL || ttl > maxUploadIntentTTL {
			return nil, apperrors.BadRequestError("ingestion.errors.invalidIntentExpiry")
		}
	}

	key := ""
	if in.StorageKey != nil {
		key = strings.TrimSpace(*in.StorageKey)
	}

	intent, err := s.storage.CreateUploadIntent(ctx, storage.UploadIntentOptions{
		Key:         key,
		ContentType: contentType,
		Metadata:    in.Metadata,
		Expiry:      ttl,
	})
	if err != nil {
		return nil, apperrors.InternalError("ingestion.errors.createUploadIntentFailed", err)
	}

	return &CreateUploadIntentOutput{
		Key:       intent.Key,
		UploadURL: intent.UploadURL,
		Method:    intent.Method,
		Headers:   intent.Headers,
		ExpiresAt: intent.ExpiresAt,
	}, nil
}

func (s *IngestionService) FinalizeUpload(ctx context.Context, in FinalizeUploadInput) (*FinalizeUploadOutput, error) {
	if !s.enabled {
		return nil, apperrors.ForbiddenError("ingestion.errors.ingestionDisabled")
	}

	if in.AccountID == uuid.Nil || in.UserID == uuid.Nil {
		return nil, apperrors.UnauthorizedError("ingestion.errors.unauthorized")
	}

	storageKey := strings.TrimSpace(in.StorageKey)
	if storageKey == "" {
		return nil, apperrors.BadRequestError("ingestion.errors.storageKeyRequired")
	}
	idempotencyKey := strings.TrimSpace(in.IdempotencyKey)
	if idempotencyKey == "" {
		return nil, apperrors.BadRequestError("ingestion.errors.idempotencyKeyRequired")
	}

	if existing, err := s.documentRepo.GetByIdempotencyKey(ctx, in.AccountID, idempotencyKey); err != nil {
		return nil, err
	} else if existing != nil {
		return &FinalizeUploadOutput{
			IngestionID: existing.ID,
			DocumentID:  existing.ID,
			EventID:     existing.EventID,
			State:       string(existing.Status),
		}, nil
	}

	exists, err := s.storage.Exists(ctx, storageKey)
	if err != nil {
		return nil, apperrors.InternalError("ingestion.errors.storageCheckFailed", err)
	}
	if !exists {
		return nil, apperrors.NotFoundErrorWithKey("ingestion.errors.uploadedObjectNotFound")
	}

	info, err := s.storage.GetInfo(ctx, storageKey)
	if err != nil {
		return nil, apperrors.InternalError("ingestion.errors.storageInfoFailed", err)
	}

	if in.SizeBytes > 0 && info.Size != in.SizeBytes {
		return nil, apperrors.BadRequestError("ingestion.errors.sizeMismatch")
	}

	if contentType := strings.TrimSpace(in.ContentType); contentType != "" {
		if infoType := strings.TrimSpace(info.ContentType); infoType != "" && !strings.EqualFold(contentType, infoType) {
			return nil, apperrors.BadRequestError("ingestion.errors.contentTypeMismatch")
		}
	}

	documentID := uuid.New()
	eventID := uuid.New()

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		doc := &entity.IngestionDocument{
			BaseModel:        model.BaseModel{ID: documentID},
			AccountID:        in.AccountID,
			UserID:           in.UserID,
			StorageKey:       storageKey,
			ContentType:      fallbackString(in.ContentType, info.ContentType, "application/octet-stream"),
			SizeBytes:        info.Size,
			ChecksumSHA256:   strings.TrimSpace(in.ChecksumSHA256),
			IdempotencyKey:   idempotencyKey,
			BatchID:          in.BatchID,
			SourceFilename:   sanitizeNullableString(in.SourceFilename),
			DeclaredLanguage: sanitizeNullableString(in.DeclaredLanguage),
			SchemaVersion:    aievent.EnvelopeSchemaVersion,
			Status:           entity.IngestionDocumentStatusQueued,
			EventID:          eventID,
			SectorIDs:        in.SectorIDs,
			TagIDs:           in.TagIDs,
			Region:           sanitizeNullableString(in.Region),
			Stage:            sanitizeNullableString(in.Stage),
		}
		if doc.ChecksumSHA256 == "" {
			doc.ChecksumSHA256 = "unverified"
		}

		if err := s.documentRepo.Create(txCtx, doc); err != nil {
			return err
		}

		innerPayload := datatypes.JSONMap{
			"document_id":            documentID.String(),
			"storage_key":            doc.StorageKey,
			"content_type":           doc.ContentType,
			"size_bytes":             doc.SizeBytes,
			"checksum_sha256":        doc.ChecksumSHA256,
			"payload_schema_version": aievent.IngestionRequestedPayloadSchemaVersion,
		}
		if doc.SourceFilename != nil {
			innerPayload["source_filename"] = *doc.SourceFilename
		}
		if doc.DeclaredLanguage != nil {
			innerPayload["declared_language"] = *doc.DeclaredLanguage
		}
		payload := datatypes.JSONMap{
			"ingestion_requested": innerPayload,
		}

		outbox := &entity.IngestionOutbox{
			EventID:        eventID,
			EventType:      aievent.DocumentIngestionRequestedV1,
			SchemaVersion:  aievent.EnvelopeSchemaVersion,
			Producer:       aievent.ProducerCoreBackend,
			KeyID:          "pending-signature",
			IdempotencyKey: idempotencyKey,
			AggregateID:    documentID,
			AccountID:      in.AccountID,
			UserID:         in.UserID,
			BatchID:        in.BatchID,
			ReplayCount:    0,
			Payload:        payload,
			Status:         entity.OutboxStatusPending,
			AttemptCount:   0,
		}
		if err := s.outboxRepo.Create(txCtx, outbox); err != nil {
			return err
		}

		now := time.Now().UTC()
		projection := &entity.IngestionStatusProjection{
			ID:                uuid.New(),
			DocumentID:        documentID,
			AccountID:         in.AccountID,
			UserID:            in.UserID,
			EventID:           eventID.String(),
			CurrentStage:      entity.IngestionStageQueued,
			IsTerminal:        false,
			StartedAt:         now,
			UpdatedAt:         now,
			LastEventSequence: now.UnixMilli(),
		}
		if err := s.projectionRepo.UpsertProjection(txCtx, projection); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.InternalError("ingestion.errors.finalizeFailed", err)
	}

	return &FinalizeUploadOutput{
		IngestionID: documentID,
		DocumentID:  documentID,
		EventID:     eventID,
		State:       string(entity.IngestionDocumentStatusQueued),
	}, nil
}

func (s *IngestionService) DeleteDocument(ctx context.Context, documentID uuid.UUID, accountID uuid.UUID) error {
	if documentID == uuid.Nil || accountID == uuid.Nil {
		return apperrors.UnauthorizedError("ingestion.errors.unauthorized")
	}

	// 1. Fetch the document to get storage key and verify ownership.
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if doc.AccountID != accountID {
		return apperrors.ForbiddenError("ingestion.errors.forbidden")
	}

	// 2. Delete the file from object storage.
	// Best-effort: continue cleanup even if storage removal fails.
	if err := s.storage.Delete(ctx, doc.StorageKey); err != nil {
		_ = err
	}

	// 3. Publish lifecycle removed event via outbox.
	eventID := uuid.New()
	payload := datatypes.JSONMap{
		"document_id":            documentID.String(),
		"storage_key":            doc.StorageKey,
		"payload_schema_version": aievent.LifecycleRemovedPayloadSchemaVersion,
	}
	outbox := &entity.IngestionOutbox{
		EventID:        eventID,
		EventType:      aievent.DocumentLifecycleRemovedV1,
		SchemaVersion:  aievent.EnvelopeSchemaVersion,
		Producer:       aievent.ProducerCoreBackend,
		KeyID:          "pending-signature",
		IdempotencyKey: "lifecycle-removed-" + documentID.String(),
		AggregateID:    documentID,
		AccountID:      doc.AccountID,
		UserID:         doc.UserID,
		BatchID:        nil,
		ReplayCount:    0,
		Payload: datatypes.JSONMap{
			"lifecycle_removed": payload,
		},
		Status:       entity.OutboxStatusPending,
		AttemptCount: 0,
	}
	if err := s.outboxRepo.Create(ctx, outbox); err != nil {
		return err
	}

	// 4. Soft-delete the document record.
	if err := s.documentRepo.SoftDelete(ctx, documentID, accountID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); !ok || appErr.GetCode() != apperrors.ErrCodeNotFound {
			return err
		}
	}

	// 5. Hard-delete the status projection.
	if err := s.projectionRepo.DeleteByDocumentID(ctx, documentID); err != nil {
		return err
	}

	return nil
}

func (s *IngestionService) GetDocument(ctx context.Context, documentID uuid.UUID, accountID uuid.UUID) (*entity.IngestionDocument, error) {
	if documentID == uuid.Nil || accountID == uuid.Nil {
		return nil, apperrors.UnauthorizedError("ingestion.errors.unauthorized")
	}
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.AccountID != accountID {
		return nil, apperrors.ForbiddenError("ingestion.errors.forbidden")
	}
	return doc, nil
}

func sanitizeNullableString(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func fallbackString(primary, secondary, fallback string) string {
	if p := strings.TrimSpace(primary); p != "" {
		return p
	}
	if s := strings.TrimSpace(secondary); s != "" {
		return s
	}
	return fallback
}
