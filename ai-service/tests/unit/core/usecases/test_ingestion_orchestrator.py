from __future__ import annotations

import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock, MagicMock, patch

import httpx
import pytest

from core.domain.enums import IngestionStage
from core.domain.exceptions import InvalidStateTransitionError
from core.ports.chunking import Chunk, ChunkProvenance
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase

FILER_URL = "http://seaweedfs-filer:8888"


def _payload() -> dict[str, object]:
    return {
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


def _make_chunks(texts: list[str]) -> list[Chunk]:
    return [
        Chunk(
            chunk_text=text,
            token_count=2,
            provenance=ChunkProvenance(
                section_heading="Intro",
                section_order=0,
                chunk_index=idx,
            ),
        )
        for idx, text in enumerate(texts)
    ]


class _Harness:
    def __init__(
        self,
        *,
        event_bus: object | None = None,
        emit_status_events: bool = True,
        seaweedfs_filer_url: str = FILER_URL,
    ) -> None:
        self.parser = AsyncMock()
        self.parser.parse = AsyncMock(return_value=SimpleNamespace())
        self.chunker = AsyncMock()
        self.chunker.chunk = AsyncMock(return_value=_make_chunks(["Hello world"]))

        self.parser_registry = AsyncMock()
        self.parser_registry.get_parser = MagicMock(return_value=self.parser)
        self.chunking_registry = AsyncMock()
        self.chunking_registry.get_chunker = MagicMock(return_value=self.chunker)

        self.embedding_port = AsyncMock()
        self.embedding_port.embed_documents = AsyncMock(return_value=[[0.1, 0.2]])
        self.knowledge_repository = AsyncMock()
        self.core_service_port = AsyncMock()

        self.use_case = IngestionOrchestratorUseCase(
            event_bus=event_bus,  # type: ignore[arg-type]
            emit_status_events=emit_status_events,
            parser_registry=self.parser_registry,  # type: ignore[arg-type]
            chunking_registry=self.chunking_registry,  # type: ignore[arg-type]
            embedding_port=self.embedding_port,  # type: ignore[arg-type]
            knowledge_repository=self.knowledge_repository,  # type: ignore[arg-type]
            core_service_port=self.core_service_port,  # type: ignore[arg-type]
            seaweedfs_filer_url=seaweedfs_filer_url,
        )


def _patch_http_client(*, status_code: int = 200):
    client = MagicMock()
    response = httpx.Response(
        status_code,
        content=b"document bytes",
        request=httpx.Request("GET", f"{FILER_URL}/"),
    )
    client.get = AsyncMock(return_value=response)
    client.__aenter__ = AsyncMock(return_value=client)
    client.__aexit__ = AsyncMock(return_value=False)
    return patch("httpx.AsyncClient", return_value=client)


@pytest.mark.asyncio
async def test_start_ingestion_runs_pipeline_to_completion() -> None:
    harness = _Harness()

    with _patch_http_client():
        result = await harness.use_case.start_ingestion(_payload())

    assert result.is_terminal is True
    assert result.status == "ok"
    assert result.context.from_stage is IngestionStage.INDEXING
    assert result.context.to_stage is IngestionStage.COMPLETED

    harness.embedding_port.embed_documents.assert_awaited_once_with(
        ["Hello world"],
        input_type="search_document",
    )
    harness.knowledge_repository.upsert_document.assert_awaited_once()
    harness.knowledge_repository.complete_ingestion_atomically.assert_awaited_once()


@pytest.mark.asyncio
async def test_start_ingestion_publishes_stage_transitions_in_order() -> None:
    event_bus = AsyncMock()
    harness = _Harness(event_bus=event_bus)

    with _patch_http_client():
        await harness.use_case.start_ingestion(_payload())

    published_stages = [
        call.args[1]["payload"]["to_stage"] for call in event_bus.publish.await_args_list
    ]
    assert published_stages == [
        IngestionStage.VALIDATING.value,
        IngestionStage.FETCHING.value,
        IngestionStage.CHUNKING.value,
        IngestionStage.EMBEDDING.value,
        IngestionStage.INDEXING.value,
        IngestionStage.COMPLETED.value,
    ]
    assert event_bus.publish.await_args_list[-1].args[1]["payload"]["is_terminal"] is True


@pytest.mark.asyncio
async def test_orchestrator_emits_status_event_on_transition() -> None:
    mock_event_bus = AsyncMock()
    harness = _Harness(event_bus=mock_event_bus)

    with _patch_http_client():
        await harness.use_case.start_ingestion(_payload())

    mock_event_bus.publish.assert_awaited()


@pytest.mark.asyncio
async def test_orchestrator_skips_event_emission_when_disabled() -> None:
    mock_event_bus = AsyncMock()
    harness = _Harness(event_bus=mock_event_bus, emit_status_events=False)

    with _patch_http_client():
        await harness.use_case.start_ingestion(_payload())

    mock_event_bus.publish.assert_not_awaited()


@pytest.mark.asyncio
async def test_start_ingestion_returns_failed_when_document_deleted() -> None:
    harness = _Harness()

    with _patch_http_client(status_code=404):
        result = await harness.use_case.start_ingestion(_payload())

    assert result.status == "failed"
    assert result.is_terminal is True
    assert result.context.to_stage is IngestionStage.FAILED
    assert result.context.from_stage is IngestionStage.VALIDATING


@pytest.mark.asyncio
async def test_start_ingestion_raises_on_stage_failure() -> None:
    harness = _Harness()
    harness.parser.parse = AsyncMock(side_effect=RuntimeError("parse exploded"))

    with _patch_http_client(), pytest.raises(RuntimeError, match="parse exploded"):
        await harness.use_case.start_ingestion(_payload())


@pytest.mark.asyncio
async def test_start_ingestion_rejects_missing_ingestion_requested() -> None:
    harness = _Harness()
    payload = _payload()
    payload["payload"] = {}

    with pytest.raises(InvalidStateTransitionError, match="missing ingestion_requested"):
        await harness.use_case.start_ingestion(payload)


@pytest.mark.asyncio
async def test_start_ingestion_rejects_missing_envelope_payload() -> None:
    harness = _Harness()
    payload = _payload()
    payload["payload"] = None

    with pytest.raises(InvalidStateTransitionError, match="missing envelope payload"):
        await harness.use_case.start_ingestion(payload)


@pytest.mark.asyncio
async def test_start_ingestion_rejects_invalid_document_id() -> None:
    harness = _Harness()
    payload = _payload()
    payload["payload"]["ingestion_requested"]["document_id"] = "not-a-uuid"  # type: ignore[index]

    with pytest.raises(InvalidStateTransitionError, match="invalid document_id"):
        await harness.use_case.start_ingestion(payload)


@pytest.mark.asyncio
async def test_start_ingestion_rejects_missing_occurred_at() -> None:
    harness = _Harness()
    payload = _payload()
    del payload["occurred_at"]

    with pytest.raises(InvalidStateTransitionError, match="missing occurred_at"):
        await harness.use_case.start_ingestion(payload)


@pytest.mark.asyncio
async def test_start_ingestion_rejects_invalid_occurred_at() -> None:
    harness = _Harness()
    payload = _payload()
    payload["occurred_at"] = "not-a-date"

    with pytest.raises(InvalidStateTransitionError, match="invalid occurred_at"):
        await harness.use_case.start_ingestion(payload)


@pytest.mark.asyncio
async def test_start_ingestion_rejects_naive_occurred_at() -> None:
    harness = _Harness()
    payload = _payload()
    payload["occurred_at"] = "2026-04-16T10:00:00"

    with pytest.raises(InvalidStateTransitionError, match="occurred_at must be timezone-aware"):
        await harness.use_case.start_ingestion(payload)


@pytest.mark.asyncio
async def test_start_ingestion_raises_when_filer_url_unconfigured() -> None:
    harness = _Harness(seaweedfs_filer_url="")

    with pytest.raises(RuntimeError, match="SEAWEEDFS_FILER_URL not configured"):
        await harness.use_case.start_ingestion(_payload())


@pytest.mark.asyncio
async def test_start_ingestion_reraises_non_not_found_http_error() -> None:
    harness = _Harness()

    with _patch_http_client(status_code=500), pytest.raises(httpx.HTTPStatusError):
        await harness.use_case.start_ingestion(_payload())


@pytest.mark.asyncio
async def test_start_ingestion_raises_on_unsupported_content_type() -> None:
    harness = _Harness()
    harness.parser_registry.get_parser = MagicMock(return_value=None)

    with _patch_http_client(), pytest.raises(ValueError, match="unsupported content type"):
        await harness.use_case.start_ingestion(_payload())


@pytest.mark.asyncio
async def test_start_ingestion_raises_when_chunker_missing() -> None:
    harness = _Harness()
    harness.chunking_registry.get_chunker = MagicMock(return_value=None)

    with _patch_http_client(), pytest.raises(ValueError, match="structural chunker not available"):
        await harness.use_case.start_ingestion(_payload())


@pytest.mark.asyncio
async def test_start_ingestion_swallows_event_publish_failure() -> None:
    event_bus = AsyncMock()
    event_bus.publish = AsyncMock(side_effect=RuntimeError("bus down"))
    harness = _Harness(event_bus=event_bus)

    with _patch_http_client():
        result = await harness.use_case.start_ingestion(_payload())

    assert result.status == "ok"
    assert result.is_terminal is True
