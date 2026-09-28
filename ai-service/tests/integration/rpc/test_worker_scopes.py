from __future__ import annotations

import asyncio
import hashlib
import hmac
import json
import uuid
from datetime import UTC, datetime
from unittest.mock import AsyncMock

import httpx
import pytest

from core.domain.enums import DocumentStatus
from core.domain.exceptions import RepositoryError
from core.domain.ingestion_events import DOCUMENT_INGESTION_REQUESTED_V1
from core.ports.embedding import EmbeddingPort
from core.ports.ingestion_event_ledger import RecordIngestionEventResult
from core.security import EnvelopeVerifier
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase
from infrastructure.chunking.registry import ChunkingRegistry
from infrastructure.database.connection import Database
from infrastructure.database.scoped import ScopedRepository
from infrastructure.messagebus import MessageHandlingRejectError
from infrastructure.parsers.registry import ParserRegistry
from workers.ingestion_worker import handle_ingestion_message
from workers.tasks.ingestion_requested import IngestionRequestedTaskHandler

TEST_SECRET = "worker-test-only"  # noqa: S105


def _envelope(document_id: uuid.UUID) -> dict[str, object]:
    envelope: dict[str, object] = {
        "event_id": str(uuid.uuid4()),
        "event_type": DOCUMENT_INGESTION_REQUESTED_V1,
        "idempotency_key": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
        "occurred_at": datetime.now(UTC).isoformat(),
        "schema_version": "1.0.0",
        "key_id": "ingestion-v1",
        "payload": {
            "ingestion_requested": {
                "document_id": str(document_id),
                "payload_schema_version": "1.0.0",
                "storage_key": "doc.txt",
                "content_type": "text/plain",
                "source_filename": "doc.txt",
                "declared_language": "en",
            }
        },
    }
    canonical = dict(envelope)
    canonical.pop("key_id")
    digest = hmac.new(
        TEST_SECRET.encode(),
        json.dumps(
            canonical,
            separators=(",", ":"),
            sort_keys=True,
        ).encode(),
        hashlib.sha256,
    ).hexdigest()
    envelope["signature"] = digest
    return envelope


@pytest.mark.integration
async def test_worker_closes_database_during_embedding_and_commits_ledger_with_index(
    database: Database,
) -> None:
    doc_id = uuid.uuid4()
    payload = _envelope(doc_id)
    started, resume = asyncio.Event(), asyncio.Event()
    embed = AsyncMock(spec=EmbeddingPort)

    async def embed_documents(*_args: object, **_kwargs: object) -> list[list[float]]:
        started.set()
        await resume.wait()
        return [[0.0] * 1024]

    embed.embed_documents.side_effect = embed_documents
    http_client = httpx.AsyncClient(
        transport=httpx.MockTransport(
            lambda _request: httpx.Response(200, content=b"Business registration guide.")
        ),
        trust_env=False,
    )
    try:
        orchestrator = IngestionOrchestratorUseCase(
            parser_registry=ParserRegistry(),
            chunking_registry=ChunkingRegistry(),
            embedding_port=embed,
            knowledge_repository=ScopedRepository(database, "knowledge"),
            core_service_port=AsyncMock(),
            seaweedfs_filer_url="http://storage.test",
            http_client=http_client,
        )
        validator = IngestionRequestedTaskHandler(
            envelope_verifier=EnvelopeVerifier(
                active_key_id="ingestion-v1", active_secret=TEST_SECRET
            ),
            ingestion_event_ledger_repository=ScopedRepository(database, "ledger"),
            ingestion_orchestrator=orchestrator,
        )
        task = asyncio.create_task(
            handle_ingestion_message(
                payload,
                database=database,
                validator=validator,
                orchestrator=orchestrator,
            )
        )
        await asyncio.wait_for(started.wait(), 5)
        assert database.engine.pool.checkedout() == 0
        resume.set()
        await asyncio.wait_for(task, 10)
        async with database.command() as reader:
            document = await reader.knowledge.get_document(doc_id)
            assert document is not None
            assert document.status is DocumentStatus.ACTIVE
            assert len(await reader.knowledge.get_chunks_by_document(doc_id)) == 1
        with pytest.raises(MessageHandlingRejectError):
            await handle_ingestion_message(
                payload,
                database=database,
                validator=validator,
                orchestrator=orchestrator,
            )
    finally:
        resume.set()
        await http_client.aclose()


@pytest.mark.integration
async def test_failed_index_rolls_back_ledger_and_document(database: Database) -> None:
    doc_id = uuid.uuid4()
    payload = _envelope(doc_id)
    embed = AsyncMock(spec=EmbeddingPort)
    embed.embed_documents.return_value = [[0.0] * 768]
    async with httpx.AsyncClient(
        transport=httpx.MockTransport(
            lambda _request: httpx.Response(200, content=b"Business guide.")
        ),
        trust_env=False,
    ) as http_client:
        orchestrator = IngestionOrchestratorUseCase(
            parser_registry=ParserRegistry(),
            chunking_registry=ChunkingRegistry(),
            embedding_port=embed,
            knowledge_repository=ScopedRepository(database, "knowledge"),
            core_service_port=AsyncMock(),
            seaweedfs_filer_url="http://storage.test",
            http_client=http_client,
        )
        validator = IngestionRequestedTaskHandler(
            envelope_verifier=EnvelopeVerifier(
                active_key_id="ingestion-v1", active_secret=TEST_SECRET
            ),
            ingestion_event_ledger_repository=ScopedRepository(database, "ledger"),
            ingestion_orchestrator=orchestrator,
        )
        with pytest.raises(RepositoryError, match="failed to complete ingestion atomically"):
            await handle_ingestion_message(
                payload,
                database=database,
                validator=validator,
                orchestrator=orchestrator,
            )
    async with database.command() as reader:
        assert await reader.knowledge.get_document(doc_id) is None
        assert (
            await reader.ledger.classify_existing(
                event_id=str(payload["event_id"]),
                idempotency_key=str(payload["idempotency_key"]),
                account_id=uuid.UUID(str(payload["account_id"])),
                document_id=doc_id,
                occurred_at=datetime.fromisoformat(str(payload["occurred_at"])),
            )
            is RecordIngestionEventResult.RECORDED
        )
