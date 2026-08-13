from __future__ import annotations

import hashlib
import hmac
import json

from app.config import Settings
from app.security import build_ingestion_envelope_verifier


def _sign(payload: dict[str, object], secret: str) -> str:
    # Mirrors the Go envelope signer (canonicalizeEnvelope): signature and
    # key_id are excluded from the canonical payload.
    body = dict(payload)
    body.pop("signature", None)
    body.pop("key_id", None)
    canonical = json.dumps(body, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return hmac.new(secret.encode("utf-8"), canonical, hashlib.sha256).hexdigest()


def test_build_ingestion_envelope_verifier_loads_previous_keys() -> None:
    settings = Settings(
        INGESTION_SIGNING_ACTIVE_KEY_ID="ingestion-v1",
        INGESTION_SIGNING_ACTIVE_KEY_SECRET="active-secret",  # noqa: S106
        INGESTION_SIGNING_PREVIOUS_KEYS_JSON='{"ingestion-v0":"old-secret"}',
    )

    verifier = build_ingestion_envelope_verifier(settings)

    envelope = {
        "event_id": "evt-1",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v0",
        "payload": {"x": 1},
    }
    envelope["signature"] = _sign(envelope, "old-secret")

    verifier.verify(envelope)


def test_build_ingestion_envelope_verifier_ignores_empty_previous_keys() -> None:
    settings = Settings(
        INGESTION_SIGNING_ACTIVE_KEY_ID="ingestion-v1",
        INGESTION_SIGNING_ACTIVE_KEY_SECRET="active-secret",  # noqa: S106
        INGESTION_SIGNING_PREVIOUS_KEYS_JSON="   ",
    )

    verifier = build_ingestion_envelope_verifier(settings)
    envelope = {
        "event_id": "evt-2",
        "event_type": "document.ingestion.requested.v1",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {"x": 1},
    }
    envelope["signature"] = _sign(envelope, "active-secret")

    verifier.verify(envelope)
