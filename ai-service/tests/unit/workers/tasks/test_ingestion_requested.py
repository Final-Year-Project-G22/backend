from __future__ import annotations

import hashlib
import hmac
import json
from datetime import UTC, datetime
from unittest.mock import AsyncMock
from uuid import uuid4

import pytest

from core.domain.ingestion_events import DOCUMENT_INGESTION_REQUESTED_V1
from core.ports.ingestion_event_ledger import RecordIngestionEventResult
from core.security import EnvelopeVerifier
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase
from infrastructure.messagebus import MessageHandlingRejectError
from workers.tasks.ingestion_requested import IngestionRequestedTaskHandler, _parse_iso_datetime

TEST_SECRET = "unit-test-secret"  # noqa: S105


def _sign(payload: dict[str, object], secret: str) -> str:
    # Mirrors the Go envelope signer (canonicalizeEnvelope): signature and
    # key_id are excluded from the canonical payload.
    body = dict(payload)
    body.pop("signature", None)
    body.pop("key_id", None)
    canonical = json.dumps(body, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return hmac.new(secret.encode("utf-8"), canonical, hashlib.sha256).hexdigest()


def _valid_envelope() -> dict[str, object]:
    envelope: dict[str, object] = {
        "event_id": "e-1",
        "event_type": DOCUMENT_INGESTION_REQUESTED_V1,
        "idempotency_key": "idem-1",
        "account_id": str(uuid4()),
        "occurred_at": "2026-04-15T12:00:00Z",
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {
            "ingestion_requested": {
                "document_id": str(uuid4()),
                "payload_schema_version": "1.0.0",
            }
        },
    }
    envelope["signature"] = _sign(envelope, TEST_SECRET)
    return envelope


@pytest.mark.asyncio
async def test_handle_accepts_valid_ingestion_requested_envelope() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    ledger = AsyncMock()
    ledger.record_or_classify.return_value = RecordIngestionEventResult.RECORDED
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )

    await handler.handle(_valid_envelope())
    ledger.record_or_classify.assert_awaited_once()
    orchestrator.start_ingestion.assert_awaited_once()


@pytest.mark.asyncio
async def test_handle_rejects_unsupported_event_type() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    ledger = AsyncMock()
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )
    envelope = _valid_envelope()
    envelope["event_type"] = "document.lifecycle.archived.v1"
    envelope["signature"] = _sign(envelope, TEST_SECRET)

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(envelope)


@pytest.mark.asyncio
async def test_handle_rejects_unsupported_payload_schema() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    ledger = AsyncMock()
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )
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
    ledger = AsyncMock()
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )
    envelope = _valid_envelope()
    envelope["signature"] = "invalid"

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(envelope)


@pytest.mark.asyncio
async def test_handle_rejects_duplicate_event_id() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    ledger = AsyncMock()
    ledger.record_or_classify.return_value = RecordIngestionEventResult.DUPLICATE_EVENT
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(_valid_envelope())


@pytest.mark.asyncio
async def test_handle_rejects_stale_lineage_event() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    ledger = AsyncMock()
    ledger.record_or_classify.return_value = RecordIngestionEventResult.STALE_LINEAGE
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )

    with pytest.raises(MessageHandlingRejectError):
        await handler.handle(_valid_envelope())


def test_handler_constructor_receives_dependencies() -> None:
    verifier = EnvelopeVerifier(active_key_id="ingestion-v1", active_secret=TEST_SECRET)
    ledger = AsyncMock()
    orchestrator = AsyncMock(spec=IngestionOrchestratorUseCase)
    handler = IngestionRequestedTaskHandler(
        envelope_verifier=verifier,
        ingestion_event_ledger_repository=ledger,
        ingestion_orchestrator=orchestrator,
    )

    assert handler is not None


def test_parse_iso_datetime_accepts_z_suffix() -> None:
    parsed = _parse_iso_datetime("2026-04-15T12:00:00Z")

    assert parsed == datetime(2026, 4, 15, 12, 0, tzinfo=UTC)
