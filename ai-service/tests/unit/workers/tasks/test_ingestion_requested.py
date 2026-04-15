from __future__ import annotations

import hashlib
import hmac
import json
from unittest.mock import Mock

import pytest

from core.domain.ingestion_events import DOCUMENT_INGESTION_REQUESTED_V1
from core.security import EnvelopeVerifier
from infrastructure.messagebus import MessageHandlingRejectError
from workers.tasks.ingestion_requested import IngestionRequestedTaskHandler

TEST_SECRET = "unit-test-secret"  # noqa: S105


def _sign(payload: dict[str, object], secret: str) -> str:
    body = dict(payload)
    body.pop("signature", None)
    canonical = json.dumps(body, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return hmac.new(secret.encode("utf-8"), canonical, hashlib.sha256).hexdigest()


def _valid_envelope() -> dict[str, object]:
    envelope: dict[str, object] = {
        "event_id": "e-1",
        "event_type": DOCUMENT_INGESTION_REQUESTED_V1,
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {
            "ingestion_requested": {
                "document_id": "d-1",
                "payload_schema_version": "1.0.0",
            }
        },
    }
    envelope["signature"] = _sign(envelope, TEST_SECRET)
    return envelope


@pytest.mark.asyncio
async def test_handle_accepts_valid_ingestion_requested_envelope() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    handler = IngestionRequestedTaskHandler(envelope_verifier=verifier)

    await handler.handle(_valid_envelope())


@pytest.mark.asyncio
async def test_handle_rejects_unsupported_event_type() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    handler = IngestionRequestedTaskHandler(envelope_verifier=verifier)
    envelope = _valid_envelope()
    envelope["event_type"] = "document.lifecycle.archived.v1"
    envelope["signature"] = _sign(envelope, TEST_SECRET)

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(envelope)


@pytest.mark.asyncio
async def test_handle_rejects_unsupported_payload_schema() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    handler = IngestionRequestedTaskHandler(envelope_verifier=verifier)
    envelope = _valid_envelope()

    payload = envelope.get("payload")
    assert isinstance(payload, dict)
    ingestion_requested = payload.get("ingestion_requested")
    assert isinstance(ingestion_requested, dict)
    ingestion_requested["payload_schema_version"] = "2.0.0"

    envelope["signature"] = _sign(envelope, TEST_SECRET)

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(envelope)


@pytest.mark.asyncio
async def test_handle_rejects_invalid_signature() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    handler = IngestionRequestedTaskHandler(envelope_verifier=verifier)
    envelope = _valid_envelope()
    envelope["signature"] = "invalid"

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(envelope)


def test_handler_constructor_receives_verifier_dependency() -> None:
    verifier = Mock(spec=EnvelopeVerifier)
    handler = IngestionRequestedTaskHandler(envelope_verifier=verifier)

    assert handler is not None
