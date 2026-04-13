package event

import (
	pb "github.com/Final-Year-Project-G22/backend/core/pb/ai/ingestion/v1"
)

type IngestionEventEnvelope struct {
	EventID        string
	EventType      string
	SchemaVersion  string
	OccurredAt     string
	Producer       string
	KeyID          string
	Signature      []byte
	IdempotencyKey string
	AccountID      string
	UserID         string
	BatchID        string
	ReplayCount    int32
	Payload        IngestionPayloadUnion
}

type IngestionPayloadUnion struct {
	IngestionRequested     *DocumentIngestionRequestedPayload
	IngestionStatusUpdated *DocumentIngestionStatusUpdatedPayload
	LifecycleArchived      *DocumentLifecycleArchivedPayload
	LifecycleRemoved       *DocumentLifecycleRemovedPayload
}

type DocumentIngestionRequestedPayload struct {
	DocumentID           string
	StorageKey           string
	ContentType          string
	SizeBytes            int64
	ChecksumSHA256       string
	SourceFilename       string
	DeclaredLanguage     string
	PayloadSchemaVersion string
}

type DocumentIngestionStatusUpdatedPayload struct {
	DocumentID           string
	IngestionID          string
	Stage                string
	Message              string
	PayloadSchemaVersion string
}

type DocumentLifecycleArchivedPayload struct {
	DocumentID           string
	PayloadSchemaVersion string
}

type DocumentLifecycleRemovedPayload struct {
	DocumentID           string
	PayloadSchemaVersion string
}

func FromProtoEnvelope(pbEnvelope *pb.IngestionEventEnvelope) *IngestionEventEnvelope {
	if pbEnvelope == nil {
		return nil
	}

	envelope := &IngestionEventEnvelope{
		EventID:        pbEnvelope.GetEventId(),
		EventType:      pbEnvelope.GetEventType(),
		SchemaVersion:  pbEnvelope.GetSchemaVersion(),
		OccurredAt:     pbEnvelope.GetOccurredAt(),
		Producer:       pbEnvelope.GetProducer(),
		KeyID:          pbEnvelope.GetKeyId(),
		Signature:      pbEnvelope.GetSignature(),
		IdempotencyKey: pbEnvelope.GetIdempotencyKey(),
		AccountID:      pbEnvelope.GetAccountId(),
		UserID:         pbEnvelope.GetUserId(),
		BatchID:        pbEnvelope.GetBatchId(),
		ReplayCount:    pbEnvelope.GetReplayCount(),
	}

	if pbPayload := pbEnvelope.GetPayload(); pbPayload != nil {
		switch p := pbPayload.Payload.(type) {
		case *pb.IngestionPayload_IngestionRequested:
			if req := p.IngestionRequested; req != nil {
				envelope.Payload.IngestionRequested = &DocumentIngestionRequestedPayload{
					DocumentID:           req.GetDocumentId(),
					StorageKey:           req.GetStorageKey(),
					ContentType:          req.GetContentType(),
					SizeBytes:            req.GetSizeBytes(),
					ChecksumSHA256:       req.GetChecksumSha256(),
					SourceFilename:       req.GetSourceFilename(),
					DeclaredLanguage:     req.GetDeclaredLanguage(),
					PayloadSchemaVersion: req.GetPayloadSchemaVersion(),
				}
			}
		case *pb.IngestionPayload_IngestionStatusUpdated:
			if status := p.IngestionStatusUpdated; status != nil {
				envelope.Payload.IngestionStatusUpdated = &DocumentIngestionStatusUpdatedPayload{
					DocumentID:           status.GetDocumentId(),
					IngestionID:          status.GetIngestionId(),
					Stage:                status.GetStage(),
					Message:              status.GetMessage(),
					PayloadSchemaVersion: status.GetPayloadSchemaVersion(),
				}
			}
		case *pb.IngestionPayload_LifecycleArchived:
			if archived := p.LifecycleArchived; archived != nil {
				envelope.Payload.LifecycleArchived = &DocumentLifecycleArchivedPayload{
					DocumentID:           archived.GetDocumentId(),
					PayloadSchemaVersion: archived.GetPayloadSchemaVersion(),
				}
			}
		case *pb.IngestionPayload_LifecycleRemoved:
			if removed := p.LifecycleRemoved; removed != nil {
				envelope.Payload.LifecycleRemoved = &DocumentLifecycleRemovedPayload{
					DocumentID:           removed.GetDocumentId(),
					PayloadSchemaVersion: removed.GetPayloadSchemaVersion(),
				}
			}
		}
	}

	return envelope
}

func ToProtoEnvelope(envelope *IngestionEventEnvelope) *pb.IngestionEventEnvelope {
	if envelope == nil {
		return nil
	}

	pbEnvelope := &pb.IngestionEventEnvelope{
		EventId:        envelope.EventID,
		EventType:      envelope.EventType,
		SchemaVersion:  envelope.SchemaVersion,
		OccurredAt:     envelope.OccurredAt,
		Producer:       envelope.Producer,
		KeyId:          envelope.KeyID,
		Signature:      envelope.Signature,
		IdempotencyKey: envelope.IdempotencyKey,
		AccountId:      envelope.AccountID,
		UserId:         envelope.UserID,
		BatchId:        envelope.BatchID,
		ReplayCount:    envelope.ReplayCount,
	}

	pbPayload := &pb.IngestionPayload{}

	if payload := envelope.Payload.IngestionRequested; payload != nil {
		pbPayload.Payload = &pb.IngestionPayload_IngestionRequested{
			IngestionRequested: &pb.DocumentIngestionRequestedPayload{
				DocumentId:           payload.DocumentID,
				StorageKey:           payload.StorageKey,
				ContentType:          payload.ContentType,
				SizeBytes:            payload.SizeBytes,
				ChecksumSha256:       payload.ChecksumSHA256,
				SourceFilename:       payload.SourceFilename,
				DeclaredLanguage:     payload.DeclaredLanguage,
				PayloadSchemaVersion: payload.PayloadSchemaVersion,
			},
		}
	} else if payload := envelope.Payload.IngestionStatusUpdated; payload != nil {
		pbPayload.Payload = &pb.IngestionPayload_IngestionStatusUpdated{
			IngestionStatusUpdated: &pb.DocumentIngestionStatusUpdatedPayload{
				DocumentId:           payload.DocumentID,
				IngestionId:          payload.IngestionID,
				Stage:                payload.Stage,
				Message:              payload.Message,
				PayloadSchemaVersion: payload.PayloadSchemaVersion,
			},
		}
	} else if payload := envelope.Payload.LifecycleArchived; payload != nil {
		pbPayload.Payload = &pb.IngestionPayload_LifecycleArchived{
			LifecycleArchived: &pb.DocumentLifecycleArchivedPayload{
				DocumentId:           payload.DocumentID,
				PayloadSchemaVersion: payload.PayloadSchemaVersion,
			},
		}
	} else if payload := envelope.Payload.LifecycleRemoved; payload != nil {
		pbPayload.Payload = &pb.IngestionPayload_LifecycleRemoved{
			LifecycleRemoved: &pb.DocumentLifecycleRemovedPayload{
				DocumentId:           payload.DocumentID,
				PayloadSchemaVersion: payload.PayloadSchemaVersion,
			},
		}
	}

	pbEnvelope.Payload = pbPayload

	return pbEnvelope
}
