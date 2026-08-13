from __future__ import annotations

import hashlib
import hmac
import json

import pytest

from core.security.envelope_verifier import (
    EnvelopeVerificationError,
    EnvelopeVerifier,
    _canonical_json,
)

TEST_ACTIVE_SECRET = "unit-test-active-secret"  # noqa: S105
TEST_PREVIOUS_SECRET = "unit-test-previous-secret"  # noqa: S105


def _sign(payload: dict[str, object], secret: str) -> str:
    # Mirrors the Go envelope signer (canonicalizeEnvelope): signature and
    # key_id are excluded from the canonical payload.
    body = dict(payload)
    body.pop("signature", None)
    body.pop("key_id", None)
    canonical = json.dumps(body, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return hmac.new(secret.encode("utf-8"), canonical, hashlib.sha256).hexdigest()


def test_verify_accepts_go_signer_golden_signature() -> None:
    # Golden vector for the Go envelope signer algorithm
    # (core-backend/internal/modules/ai/domain/service/envelope_signer.go):
    # canonical JSON = envelope minus signature/key_id, sorted keys, compact
    # separators, ensure_ascii. The canonical bytes below are exactly what the
    # Go canonicalizer emits for this ASCII payload, and the signature is
    # HMAC-SHA256(secret, canonical) — deterministic in both languages.
    # (Non-ASCII strings and floats may still diverge between Go and Python;
    # those are covered separately if payloads ever carry them.)
    envelope = {
        "event_id": "e-1",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {"x": 1},
    }
    expected_canonical = (
        b'{"event_id":"e-1","event_type":"document.ingestion.requested.v1",'
        b'"payload":{"x":1},"schema_version":"1.0.0"}'
    )
    expected_signature = "8e3aa89d4296b3a7902574c051edda20b70167cf688246fd1cc6096f78e83d6c"
    envelope["signature"] = expected_signature

    assert (
        _canonical_json({k: v for k, v in envelope.items() if k not in ("signature", "key_id")})
        == expected_canonical
    )
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_ACTIVE_SECRET)
    verifier.verify(envelope)


def test_verify_accepts_active_key_signature() -> None:
    secret = TEST_ACTIVE_SECRET
    envelope = {
        "event_id": "e-1",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {"x": 1},
    }
    envelope["signature"] = _sign(envelope, secret)

    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=secret)
    verifier.verify(envelope)


def test_verify_accepts_previous_key_signature() -> None:
    prev_secret = TEST_PREVIOUS_SECRET
    envelope = {
        "event_id": "e-2",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v0",
        "payload": {"x": 1},
    }
    envelope["signature"] = _sign(envelope, prev_secret)

    verifier = EnvelopeVerifier(
        active_key_id="ingestion-v1",
        active_secret=TEST_ACTIVE_SECRET,
        previous_keys={"ingestion-v0": prev_secret},
    )
    verifier.verify(envelope)


def test_verify_rejects_unknown_key() -> None:
    envelope = {
        "event_id": "e-3",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "unknown",
        "payload": {"x": 1},
        "signature": "abc",
    }
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_ACTIVE_SECRET)

    with pytest.raises(EnvelopeVerificationError):
        verifier.verify(envelope)


def test_verify_rejects_invalid_signature() -> None:
    envelope = {
        "event_id": "e-4",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {"x": 1},
        "signature": "not-valid",
    }
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_ACTIVE_SECRET)

    with pytest.raises(EnvelopeVerificationError):
        verifier.verify(envelope)


def test_verify_rejects_schema_mismatch() -> None:
    envelope = {
        "event_id": "e-5",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "2.0.0",
        "key_id": "ingestion-v1",
        "payload": {"x": 1},
        "signature": "abc",
    }
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_ACTIVE_SECRET)

    with pytest.raises(EnvelopeVerificationError):
        verifier.verify(envelope)
