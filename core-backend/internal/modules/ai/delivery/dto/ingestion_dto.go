package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateUploadIntentRequest struct {
	StorageKey   *string           `json:"storageKey,omitempty" doc:"Optional object key. If omitted, server generates one"`
	ContentType  string            `json:"contentType" doc:"Expected MIME content type" minLength:"1" maxLength:"255"`
	Metadata     map[string]string `json:"metadata,omitempty" doc:"Optional metadata sent as upload headers"`
	ExpiresInSec *int              `json:"expiresInSec,omitempty" doc:"Optional upload intent expiration in seconds" minimum:"60" maximum:"3600"`
	BatchID      *uuid.UUID        `json:"batchId,omitempty" doc:"Optional client batch identifier"`
}

type CreateUploadIntentInput struct {
	Body CreateUploadIntentRequest
}

type CreateUploadIntentResponseBody struct {
	Key       string            `json:"key" doc:"Resolved object key for upload"`
	UploadURL string            `json:"uploadUrl" doc:"Direct upload URL"`
	Method    string            `json:"method" doc:"HTTP method for upload"`
	Headers   map[string]string `json:"headers,omitempty" doc:"Headers required by storage provider"`
	ExpiresAt time.Time         `json:"expiresAt" doc:"Upload intent expiry timestamp"`
}

type CreateUploadIntentOutput struct {
	Body CreateUploadIntentResponseBody
}

type FinalizeUploadRequest struct {
	StorageKey       string     `json:"storageKey" doc:"Uploaded object key" minLength:"1"`
	ContentType      string     `json:"contentType" doc:"Uploaded object MIME type" minLength:"1" maxLength:"255"`
	SizeBytes        int64      `json:"sizeBytes" doc:"Uploaded object size in bytes" minimum:"0"`
	ChecksumSHA256   string     `json:"checksumSha256" doc:"Hex encoded SHA-256 checksum" minLength:"1" maxLength:"128"`
	IdempotencyKey   string     `json:"idempotencyKey" doc:"Client idempotency key for dedupe" minLength:"1" maxLength:"255"`
	SourceFilename   *string    `json:"sourceFilename,omitempty" doc:"Optional original filename"`
	DeclaredLanguage *string    `json:"declaredLanguage,omitempty" doc:"Optional language hint (e.g. en, am)"`
	BatchID          *uuid.UUID `json:"batchId,omitempty" doc:"Optional batch identifier"`
}

type FinalizeUploadInput struct {
	Body FinalizeUploadRequest
}

type FinalizeUploadResponseBody struct {
	IngestionID uuid.UUID `json:"ingestionId" doc:"Ingestion document identifier"`
	DocumentID  uuid.UUID `json:"documentId" doc:"Alias of ingestion document identifier"`
	EventID     uuid.UUID `json:"eventId" doc:"Outbox event identifier"`
	State       string    `json:"state" doc:"Initial ingestion state"`
}

type FinalizeUploadOutput struct {
	Body FinalizeUploadResponseBody
}
