from __future__ import annotations

import uuid
from unittest.mock import AsyncMock

import pytest

from core.domain.enums import IngestionStage
from core.domain.exceptions import InvalidStateTransitionError
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase


def _payload(*, current_stage: str | None = None) -> dict[str, object]:
    payload: dict[str, object] = {
        "event_id": "evt-1",
        "event_type": "document.ingestion.requested.v1",
        "idempotency_key": "idem-1",
        "account_id": str(uuid.uuid4()),
        "occurred_at": "2026-04-16T10:00:00Z",
        "schema_version": "1.0.0",
        "producer": "core-backend",
        "replay_count": 0,
        "payload": {
            "ingestion_requested": {
                "document_id": str(uuid.uuid4()),
            }
        },
    }
    if current_stage is not None:
        payload["current_stage"] = current_stage
    return payload


@pytest.mark.asyncio
async def test_start_ingestion_moves_none_to_queued() -> None:
    use_case = IngestionOrchestratorUseCase()

    result = await use_case.start_ingestion(_payload())

    assert result.context.from_stage is None
    assert result.context.to_stage is IngestionStage.QUEUED
    assert result.is_terminal is False


@pytest.mark.asyncio
async def test_start_ingestion_advances_linear_stage() -> None:
    use_case = IngestionOrchestratorUseCase()

    result = await use_case.start_ingestion(_payload(current_stage=IngestionStage.CHUNKING.value))

    assert result.context.from_stage is IngestionStage.CHUNKING
    assert result.context.to_stage is IngestionStage.EMBEDDING


@pytest.mark.asyncio
async def test_start_ingestion_rejects_transition_from_terminal_stage() -> None:
    use_case = IngestionOrchestratorUseCase()

    with pytest.raises(InvalidStateTransitionError):
        await use_case.start_ingestion(_payload(current_stage=IngestionStage.COMPLETED.value))


@pytest.mark.asyncio
async def test_start_ingestion_rejects_invalid_stage_value() -> None:
    use_case = IngestionOrchestratorUseCase()

    with pytest.raises(InvalidStateTransitionError):
        await use_case.start_ingestion(_payload(current_stage="broken"))


@pytest.mark.asyncio
async def test_orchestrator_emits_status_event_on_transition() -> None:
    mock_event_bus = AsyncMock()
    use_case = IngestionOrchestratorUseCase(event_bus=mock_event_bus, emit_status_events=True)

    payload = _payload()
    await use_case.start_ingestion(payload)

    mock_event_bus.publish.assert_awaited_once()


@pytest.mark.asyncio
async def test_orchestrator_skips_event_emission_when_disabled() -> None:
    mock_event_bus = AsyncMock()
    use_case = IngestionOrchestratorUseCase(event_bus=mock_event_bus, emit_status_events=False)

    payload = _payload()
    await use_case.start_ingestion(payload)

    mock_event_bus.publish.assert_not_awaited()


@pytest.mark.asyncio
async def test_terminal_stage_sets_is_terminal_true() -> None:
    use_case = IngestionOrchestratorUseCase()

    result = await use_case.start_ingestion(_payload(current_stage=IngestionStage.INDEXING.value))

    assert result.is_terminal is True
