from __future__ import annotations

import logging
from datetime import datetime
from typing import Any, cast
from uuid import UUID

from core.domain.exceptions import RepositoryError
from core.domain.ingestion_events import (
    ACCEPTED_INGESTION_REQUESTED_SCHEMA_VERSIONS,
    DOCUMENT_INGESTION_REQUESTED_V1,
)
from core.ports.ingestion_event_ledger import IngestionEventLedgerPort, RecordIngestionEventResult
from core.security import EnvelopeVerificationError, EnvelopeVerifier
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase
from infrastructure.messagebus import MessageHandlingRejectError, MessageHandlingRetryError

logger = logging.getLogger(__name__)


class IngestionRequestedTaskHandler:
    def __init__(
        self,
        *,
        envelope_verifier: EnvelopeVerifier,
        ingestion_event_ledger_repository: IngestionEventLedgerPort,
        ingestion_orchestrator: IngestionOrchestratorUseCase,
    ) -> None:
        self._envelope_verifier = envelope_verifier
        self._ingestion_event_ledger_repository = ingestion_event_ledger_repository
        self._ingestion_orchestrator = ingestion_orchestrator

    async def handle(self, payload: dict[str, Any]) -> None:
        logger.debug(
            "event_type=%s key_id=%s has_signature=%s",
            payload.get("event_type"),
            payload.get("key_id"),
            "signature" in payload,
        )
        try:
            self._envelope_verifier.verify(payload)
        except EnvelopeVerificationError as exc:
            logger.warning(
                "envelope verification failed: %s (key_id=%s schema_version=%s has_signature=%s)",
                exc,
                payload.get("key_id"),
                payload.get("schema_version"),
                "signature" in payload,
            )
            raise MessageHandlingRejectError("invalid ingestion envelope") from exc

        event_type = str(payload.get("event_type", ""))
        if event_type != DOCUMENT_INGESTION_REQUESTED_V1:
            raise MessageHandlingRejectError("unsupported event type")

        payload_message = payload.get("payload")
        if not isinstance(payload_message, dict):
            raise MessageHandlingRejectError("missing envelope payload")
        payload_message_dict = cast(dict[str, Any], payload_message)

        ingestion_requested_payload = payload_message_dict.get("ingestion_requested")
        if not isinstance(ingestion_requested_payload, dict):
            raise MessageHandlingRejectError("missing ingestion_requested payload")
        ingestion_requested_payload_dict = cast(dict[str, Any], ingestion_requested_payload)

        payload_schema_version = str(
            ingestion_requested_payload_dict.get("payload_schema_version", "")
        )
        if payload_schema_version not in ACCEPTED_INGESTION_REQUESTED_SCHEMA_VERSIONS:
            raise MessageHandlingRejectError("unsupported ingestion payload schema")

        event_id = str(payload.get("event_id", ""))
        idempotency_key = str(payload.get("idempotency_key", ""))
        account_id = _read_uuid(payload, "account_id")
        occurred_at = _read_occurred_at(payload)

        document_id_raw = ingestion_requested_payload_dict.get("document_id")
        document_id = _parse_uuid(str(document_id_raw), field_name="document_id")

        try:
            ledger_result = await self._ingestion_event_ledger_repository.record_or_classify(
                event_id=event_id,
                idempotency_key=idempotency_key,
                account_id=account_id,
                document_id=document_id,
                occurred_at=occurred_at,
            )
        except RepositoryError as exc:
            raise MessageHandlingRetryError("failed to classify ingestion event") from exc

        if ledger_result is RecordIngestionEventResult.DUPLICATE_EVENT:
            raise MessageHandlingRejectError("duplicate event id")
        if ledger_result is RecordIngestionEventResult.DUPLICATE_IDEMPOTENCY:
            raise MessageHandlingRejectError("duplicate idempotency key")
        if ledger_result is RecordIngestionEventResult.STALE_LINEAGE:
            raise MessageHandlingRejectError("stale ingestion lineage event")

        await self._ingestion_orchestrator.start_ingestion(payload)

        logger.info(
            "ingestion event received event_id=%s event_type=%s",
            event_id,
            event_type,
        )


def _read_uuid(payload: dict[str, Any], field_name: str) -> UUID:
    raw = payload.get(field_name)
    if raw is None:
        raise MessageHandlingRejectError(f"missing {field_name}")
    return _parse_uuid(str(raw), field_name=field_name)


def _parse_uuid(value: str, *, field_name: str) -> UUID:
    try:
        return UUID(value)
    except ValueError as exc:
        raise MessageHandlingRejectError(f"invalid {field_name}") from exc


def _read_occurred_at(payload: dict[str, Any]) -> datetime:
    raw = payload.get("occurred_at")
    if raw is None:
        raise MessageHandlingRejectError("missing occurred_at")
    occurred_at = _parse_iso_datetime(str(raw))
    if occurred_at.tzinfo is None:
        raise MessageHandlingRejectError("occurred_at must be timezone-aware")
    return occurred_at


def _parse_iso_datetime(value: str) -> datetime:
    normalized = value.replace("Z", "+00:00")
    try:
        return datetime.fromisoformat(normalized)
    except ValueError as exc:
        raise MessageHandlingRejectError("invalid occurred_at") from exc


__all__ = ["IngestionRequestedTaskHandler"]
