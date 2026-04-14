package service

import (
	"context"
	"testing"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
)

func TestEnvelopeSigner_SignEnvelope(t *testing.T) {
	signer := NewEnvelopeSigner(&core.Config{
		Ingestion: core.IngestionConfig{
			Signing: core.IngestionSigningConfig{
				ActiveKeyID:     "ingestion-v1",
				ActiveKeySecret: "top-secret",
			},
		},
	})

	envelope := map[string]any{
		"event_id":        "11111111-1111-1111-1111-111111111111",
		"event_type":      "document.ingestion.requested.v1",
		"idempotency_key": "idem-123",
		"payload": map[string]any{
			"document_id": "22222222-2222-2222-2222-222222222222",
		},
	}

	sig, keyID, err := signer.SignEnvelope(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected signing error: %v", err)
	}
	if keyID != "ingestion-v1" {
		t.Fatalf("expected keyID ingestion-v1, got %s", keyID)
	}
	if len(sig) == 0 {
		t.Fatalf("expected non-empty signature")
	}
}

func TestEnvelopeSigner_MissingConfig(t *testing.T) {
	signer := NewEnvelopeSigner(&core.Config{})
	_, _, err := signer.SignEnvelope(context.Background(), map[string]any{"event_id": "x"})
	if err == nil {
		t.Fatalf("expected error for missing signing config")
	}
}
