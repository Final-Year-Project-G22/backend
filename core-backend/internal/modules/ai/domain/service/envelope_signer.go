package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	aievent "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/event"
)

type EnvelopeSigner interface {
	SignEnvelope(ctx context.Context, envelope map[string]any) (signature []byte, keyID string, err error)
}

type envelopeSigner struct {
	cfg *core.Config
}

func NewEnvelopeSigner(cfg *core.Config) EnvelopeSigner {
	return &envelopeSigner{cfg: cfg}
}

func (s *envelopeSigner) SignEnvelope(_ context.Context, envelope map[string]any) ([]byte, string, error) {
	keyID := s.cfg.Ingestion.Signing.ActiveKeyID
	secret := s.cfg.Ingestion.Signing.ActiveKeySecret
	if keyID == "" || secret == "" {
		return nil, "", fmt.Errorf("ingestion signing configuration is incomplete")
	}

	canonical, err := canonicalizeEnvelope(envelope)
	if err != nil {
		return nil, "", err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(canonical); err != nil {
		return nil, "", fmt.Errorf("failed to hash envelope: %w", err)
	}

	return []byte(hex.EncodeToString(mac.Sum(nil))), keyID, nil
}

func canonicalizeEnvelope(envelope map[string]any) ([]byte, error) {
	clone := map[string]any{}
	for k, v := range envelope {
		clone[k] = v
	}

	delete(clone, "signature")
	if _, ok := clone["schema_version"]; !ok {
		clone["schema_version"] = aievent.EnvelopeSchemaVersion
	}

	b, err := json.Marshal(clone)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize envelope: %w", err)
	}

	return b, nil
}
