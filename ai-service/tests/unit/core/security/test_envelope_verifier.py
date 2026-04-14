from __future__ import annotations

import hashlib
import hmac
import json

import pytest

from core.security.envelope_verifier import EnvelopeVerificationError, EnvelopeVerifier

TEST_ACTIVE_SECRET = "unit-test-active-secret"  # noqa: S105
TEST_PREVIOUS_SECRET = "unit-test-previous-secret"  # noqa: S105


def _sign(payload: dict[str, object], secret: str) -> str:
    body = dict(payload)
    body.pop("signature", None)
    canonical = json.dumps(body, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return hmac.new(secret.encode("utf-8"), canonical, hashlib.sha256).hexdigest()


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
