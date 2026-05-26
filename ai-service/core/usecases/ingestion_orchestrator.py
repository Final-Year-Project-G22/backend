from __future__ import annotations

import logging
import uuid
from datetime import UTC, datetime
from typing import Any, cast

import httpx

from core.domain.enums import ChunkStatus, DocumentSource, DocumentStatus, IngestionStage, Language
from core.domain.exceptions import InvalidStateTransitionError
from core.domain.ingestion_events import DOCUMENT_INGESTION_STATUS_UPDATED_V1
from core.domain.ingestion_status_events import build_status_updated_payload
from core.domain.models import DocumentChunk, KnowledgeDocument
from core.domain.value_objects import IngestionTransitionContext, IngestionTransitionResult
from core.ports.chunking import ChunkingStrategy
from core.ports.core_service import CoreServicePort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from infrastructure.chunking.registry import ChunkingRegistry
from infrastructure.parsers.registry import ParserRegistry

logger = logging.getLogger(__name__)

_TERMINAL_STAGES = frozenset(
    [
        IngestionStage.COMPLETED,
        IngestionStage.FAILED,
        IngestionStage.CANCELLED,
    ]
)

_PIPELINE_STAGES: list[IngestionStage] = [
    IngestionStage.FETCHING,
    IngestionStage.CHUNKING,
    IngestionStage.EMBEDDING,
    IngestionStage.INDEXING,
]


class IngestionOrchestratorUseCase:
    def __init__(
        self,
        *,
        event_bus: EventBusPort | None = None,
        emit_status_events: bool = True,
        parser_registry: ParserRegistry,
        chunking_registry: ChunkingRegistry,
        embedding_port: EmbeddingPort,
        knowledge_repository: KnowledgeRepositoryPort,
        core_service_port: CoreServicePort,
        seaweedfs_filer_url: str = "",
    ) -> None:
        self._event_bus = event_bus
        self._emit_status_events = emit_status_events
        self._parser_registry = parser_registry
        self._chunking_registry = chunking_registry
        self._embedding_port = embedding_port
        self._knowledge_repository = knowledge_repository
        self._core_service_port = core_service_port
        self._seaweedfs_filer_url = seaweedfs_filer_url.rstrip("/")

    async def start_ingestion(self, payload: dict[str, Any]) -> IngestionTransitionResult:
        logger.info("start_ingestion called event_id=%s", payload.get("event_id", "unknown"))
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

        content_type = str(ingestion_requested_dict.get("content_type", ""))
        source_filename = str(ingestion_requested_dict.get("source_filename", "Untitled Document"))
        storage_key = str(ingestion_requested_dict.get("storage_key", ""))
        declared_language = str(ingestion_requested_dict.get("declared_language", "en"))

        # Publish QUEUED → VALIDATING transition
        await self._publish_status_event(
            event_id=event_id,
            document_id=document_id,
            account_id=account_id,
            from_stage=None,
            to_stage=IngestionStage.VALIDATING,
            is_terminal=False,
            retry_count=0,
        )

        # Run the full pipeline: FETCHING → CHUNKING → EMBEDDING → INDEXING → COMPLETED
        current_stage = IngestionStage.VALIDATING
        document_bytes: bytes = b""
        parsed_chunks: list[Any] = []
        chunk_texts: list[str] = []
        embeddings: list[list[float]] = []

        for stage in _PIPELINE_STAGES:
            try:
                if stage == IngestionStage.FETCHING:
                    document_bytes = await self._fetch_document(document_id, storage_key)
                elif stage == IngestionStage.CHUNKING:
                    parsed_chunks = await self._parse_and_chunk(
                        document_bytes, content_type, source_filename
                    )
                    chunk_texts = [c.chunk_text for c in parsed_chunks]
                elif stage == IngestionStage.EMBEDDING:
                    embeddings = await self._embed_chunks(chunk_texts)
                elif stage == IngestionStage.INDEXING:
                    await self._index_chunks(
                        document_id,
                        source_filename,
                        content_type,
                        declared_language,
                        parsed_chunks,
                        embeddings,
                    )

                await self._publish_status_event(
                    event_id=event_id,
                    document_id=document_id,
                    account_id=account_id,
                    from_stage=current_stage,
                    to_stage=stage,
                    is_terminal=False,
                    retry_count=0,
                )
                current_stage = stage

            except Exception as exc:
                logger.exception(
                    "ingestion stage failed document_id=%s stage=%s",
                    document_id,
                    stage.value,
                )
                error_message = f"{stage.value}: {exc}"
                await self._publish_status_event(
                    event_id=event_id,
                    document_id=document_id,
                    account_id=account_id,
                    from_stage=current_stage,
                    to_stage=IngestionStage.FAILED,
                    is_terminal=True,
                    retry_count=0,
                    error_message=error_message,
                )
                raise

        # Pipeline completed successfully
        await self._publish_status_event(
            event_id=event_id,
            document_id=document_id,
            account_id=account_id,
            from_stage=current_stage,
            to_stage=IngestionStage.COMPLETED,
            is_terminal=True,
            retry_count=0,
        )

        logger.info("ingestion completed document_id=%s", document_id)
        return IngestionTransitionResult(
            context=IngestionTransitionContext(
                event_id=event_id,
                document_id=document_id,
                account_id=account_id,
                idempotency_key=idempotency_key,
                from_stage=current_stage,
                to_stage=IngestionStage.COMPLETED,
                occurred_at=occurred_at,
                retry_count=0,
                metadata={},
            ),
            is_terminal=True,
            status="ok",
        )

    async def _fetch_document(
        self,
        document_id: uuid.UUID,
        storage_key: str,
    ) -> bytes:
        """Fetch document bytes directly from SeaweedFS filer."""
        logger.info(
            "fetching document document_id=%s storage_key=%s",
            document_id,
            storage_key,
        )

        if not self._seaweedfs_filer_url:
            raise RuntimeError("SEAWEEDFS_FILER_URL not configured")

        url = f"{self._seaweedfs_filer_url}/{storage_key}"
        async with httpx.AsyncClient() as client:
            response = await client.get(url)
            response.raise_for_status()
            logger.info(
                "document downloaded document_id=%s size=%d",
                document_id,
                len(response.content),
            )
            return response.content

    async def _parse_and_chunk(
        self,
        document_bytes: bytes,
        content_type: str,
        source_filename: str,
    ) -> list[Any]:
        """Parse document bytes and split into chunks."""
        logger.info("parsing document content_type=%s", content_type)

        parser = self._parser_registry.get_parser(content_type)
        if parser is None:
            raise ValueError(f"unsupported content type: {content_type}")

        parsed = await parser.parse(document_bytes, metadata={"filename": source_filename})

        chunker = self._chunking_registry.get_chunker("structural")
        if chunker is None:
            raise ValueError("structural chunker not available")

        strategy = ChunkingStrategy(max_tokens=512, overlap_tokens=50)
        chunks = await chunker.chunk(parsed, strategy, metadata={"filename": source_filename})

        logger.info("document chunked chunks=%d", len(chunks))
        return chunks

    async def _embed_chunks(self, chunk_texts: list[str]) -> list[list[float]]:
        """Generate embeddings for chunks in batches (Cohere limit = 96)."""
        logger.info("embedding chunks count=%d", len(chunk_texts))
        batch_size = 96
        all_embeddings: list[list[float]] = []
        for i in range(0, len(chunk_texts), batch_size):
            batch = chunk_texts[i : i + batch_size]
            batch_embeddings = await self._embedding_port.embed_documents(
                batch,
                input_type="search_document",
            )
            all_embeddings.extend(batch_embeddings)
            logger.info(
                "embedded batch %d/%d count=%d",
                i // batch_size + 1,
                (len(chunk_texts) + batch_size - 1) // batch_size,
                len(batch_embeddings),
            )
        logger.info("embeddings generated total=%d", len(all_embeddings))
        return all_embeddings

    async def _index_chunks(
        self,
        document_id: uuid.UUID,
        source_filename: str,
        content_type: str,
        declared_language: str,
        chunks: list[Any],
        embeddings: list[list[float]],
    ) -> None:
        """Persist chunks with embeddings to vector store."""
        logger.info("indexing chunks document_id=%s count=%d", document_id, len(chunks))

        # Create or update the knowledge document record (sanitize null bytes)
        doc = KnowledgeDocument(
            id=document_id,
            title=source_filename.replace("\x00", ""),
            source=DocumentSource.GUIDE,
            external_id=str(document_id),
            effective_date=datetime.now(UTC).date(),
            language=Language.ENGLISH if declared_language.startswith("en") else Language.AMHARIC,
            status=DocumentStatus.PROCESSING,
            metadata={"content_type": content_type},
        )
        await self._knowledge_repository.upsert_document(doc)

        # Build document chunk entities (sanitize null bytes for PostgreSQL)
        document_chunks: list[DocumentChunk] = []
        for idx, (chunk, embedding) in enumerate(zip(chunks, embeddings, strict=True)):
            clean_text = chunk.chunk_text.replace("\x00", "")
            section_heading = (chunk.provenance.section_heading or "").replace("\x00", "")
            document_chunks.append(
                DocumentChunk(
                    document_id=document_id,
                    content_type=DocumentSource.GUIDE,
                    language=doc.language,
                    chunk_text=clean_text,
                    chunk_index=idx,
                    token_count=chunk.token_count,
                    embedding=embedding,
                    status=ChunkStatus.EMBEDDED,
                    section_heading=section_heading or None,
                    metadata={"content_type": content_type},
                )
            )

        # Atomically upsert chunks and mark document active
        await self._knowledge_repository.complete_ingestion_atomically(
            document_id,
            status=DocumentStatus.ACTIVE,
            chunks=document_chunks,
        )

        logger.info("indexing complete document_id=%s", document_id)

    async def _publish_status_event(
        self,
        *,
        event_id: str,
        document_id: uuid.UUID,
        account_id: uuid.UUID,
        from_stage: IngestionStage | None,
        to_stage: IngestionStage,
        is_terminal: bool,
        retry_count: int,
        error_message: str | None = None,
    ) -> None:
        if self._event_bus is None:
            logger.warning("event bus is None, skipping status publish document_id=%s", document_id)
            return

        status_payload = build_status_updated_payload(
            document_id=str(document_id),
            account_id=str(account_id),
            from_stage=from_stage,
            to_stage=to_stage,
            is_terminal=is_terminal,
            retry_count=retry_count,
            error_message=error_message,
        )

        envelope = {
            "event_id": f"{event_id}-status-{to_stage.value}",
            "event_type": DOCUMENT_INGESTION_STATUS_UPDATED_V1,
            "schema_version": "1.0.0",
            "payload": status_payload,
            "occurred_at": datetime.now(tz=UTC).isoformat(),
        }

        try:
            await self._event_bus.publish(
                DOCUMENT_INGESTION_STATUS_UPDATED_V1,
                envelope,
            )
            logger.info(
                "status event published document_id=%s to_stage=%s is_terminal=%s",
                document_id,
                to_stage.value,
                is_terminal,
            )
        except Exception:
            logger.exception("failed to publish status event document_id=%s", document_id)


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


__all__ = ["IngestionOrchestratorUseCase"]
