from __future__ import annotations

import uuid
from datetime import datetime
from typing import Any, cast

from core.domain.enums import IngestionStage
from core.domain.exceptions import InvalidStateTransitionError
from core.domain.value_objects import IngestionTransitionContext, IngestionTransitionResult

_NEXT_STAGE_BY_FROM: dict[IngestionStage | None, IngestionStage] = {
    None: IngestionStage.QUEUED,
    IngestionStage.QUEUED: IngestionStage.VALIDATING,
    IngestionStage.VALIDATING: IngestionStage.FETCHING,
    IngestionStage.FETCHING: IngestionStage.CHUNKING,
    IngestionStage.CHUNKING: IngestionStage.EMBEDDING,
    IngestionStage.EMBEDDING: IngestionStage.INDEXING,
    IngestionStage.INDEXING: IngestionStage.COMPLETED,
}

_TERMINAL_STAGES = frozenset(
    [
        IngestionStage.COMPLETED,
        IngestionStage.FAILED,
        IngestionStage.CANCELLED,
    ]
)


class IngestionOrchestratorUseCase:
    async def start_ingestion(self, payload: dict[str, Any]) -> IngestionTransitionResult:
        event_id = str(payload.get("event_id", ""))
        idempotency_key = str(payload.get("idempotency_key", ""))
        account_id = _read_uuid(payload, "account_id")
        occurred_at = _read_occurred_at(payload)

        payload_message = payload.get("payload")
        if not isinstance(payload_message, dict):
            raise InvalidStateTransitionError("missing envelope payload")
        payload_message_dict = cast(dict[str, Any], payload_message)

        ingestion_requested = payload_message_dict.get("ingestion_requested")
        if not isinstance(ingestion_requested, dict):
            raise InvalidStateTransitionError("missing ingestion_requested payload")
        ingestion_requested_dict = cast(dict[str, Any], ingestion_requested)

        document_id_raw = ingestion_requested_dict.get("document_id")
        document_id = _parse_uuid(str(document_id_raw), field_name="document_id")

        from_stage = _read_stage(payload.get("current_stage"))
        to_stage = self._next_stage(from_stage)

        context = IngestionTransitionContext(
            event_id=event_id,
            document_id=document_id,
            account_id=account_id,
            idempotency_key=idempotency_key,
            from_stage=from_stage,
            to_stage=to_stage,
            occurred_at=occurred_at,
            retry_count=_read_retry_count(payload.get("replay_count")),
            metadata={
                "event_type": str(payload.get("event_type", "")),
                "schema_version": str(payload.get("schema_version", "")),
                "producer": str(payload.get("producer", "")),
            },
        )

        return IngestionTransitionResult(
            context=context,
            is_terminal=to_stage in _TERMINAL_STAGES,
            status="ok",
        )

    def _next_stage(self, from_stage: IngestionStage | None) -> IngestionStage:
        if from_stage in _TERMINAL_STAGES:
            if from_stage is None:
                raise InvalidStateTransitionError("invalid terminal stage")
            raise InvalidStateTransitionError(
                "cannot transition from terminal stage",
                details={"from_stage": from_stage.value},
            )

        next_stage = _NEXT_STAGE_BY_FROM.get(from_stage)
        if next_stage is None:
            raise InvalidStateTransitionError(
                "unknown ingestion stage",
                details={"from_stage": from_stage.value if from_stage else None},
            )

        return next_stage


def _read_uuid(payload: dict[str, Any], field_name: str) -> uuid.UUID:
    raw = payload.get(field_name)
    return _parse_uuid(str(raw), field_name=field_name)


def _parse_uuid(value: str, *, field_name: str) -> uuid.UUID:
    try:
        return uuid.UUID(value)
    except ValueError as exc:
        raise InvalidStateTransitionError(f"invalid {field_name}") from exc


def _read_occurred_at(payload: dict[str, Any]) -> str:
    raw = payload.get("occurred_at")
    if raw is None:
        raise InvalidStateTransitionError("missing occurred_at")

    value = str(raw)
    normalized = value.replace("Z", "+00:00")
    try:
        parsed = datetime.fromisoformat(normalized)
    except ValueError as exc:
        raise InvalidStateTransitionError("invalid occurred_at") from exc

    if parsed.tzinfo is None:
        raise InvalidStateTransitionError("occurred_at must be timezone-aware")

    return value


def _read_stage(raw: Any) -> IngestionStage | None:
    if raw is None:
        return None
    try:
        return IngestionStage(str(raw))
    except ValueError as exc:
        raise InvalidStateTransitionError("invalid current_stage") from exc


def _read_retry_count(raw: Any) -> int:
    if raw is None:
        return 0
    try:
        value = int(raw)
    except (TypeError, ValueError) as exc:
        raise InvalidStateTransitionError("invalid replay_count") from exc

    if value < 0:
        raise InvalidStateTransitionError("replay_count must be >= 0")
    return value


__all__ = ["IngestionOrchestratorUseCase"]
