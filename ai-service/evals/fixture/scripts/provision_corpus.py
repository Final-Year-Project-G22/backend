"""Provision the evaluation fixture knowledge corpus into the ai-service DB.

Idempotent: document and chunk IDs are derived deterministically from the
manifest document keys (uuid5), so re-running replaces/upserts the same rows.
Embeddings are computed through the configured EmbeddingPort (provider from
settings); chunks are marked EMBEDDED and documents ACTIVE.

Usage (from ai-service/):
    uv run python evals/fixture/scripts/provision_corpus.py [--dry-run]
"""

from __future__ import annotations

import asyncio
import sys
import uuid
from datetime import date, datetime
from pathlib import Path

# Ensure app code is importable when run from the ai-service directory
sys.path.insert(0, str(Path(__file__).resolve().parents[3]))

from core.domain.enums import ChunkStatus, DocumentSource, DocumentStatus, Language, Tier
from core.domain.models import AIUserQuota, DocumentChunk, KnowledgeDocument
from core.ports.chunking import ChunkingStrategy
from infrastructure.chunking.registry import ChunkingRegistry
from infrastructure.database.connection import async_session_factory
from infrastructure.database.repositories.knowledge_repository import SqlAlchemyKnowledgeRepository
from infrastructure.database.repositories.quota_repository import SqlAlchemyQuotaRepository
from infrastructure.embeddings import create_embedding_adapter
from infrastructure.parsers.registry import ParserRegistry

FIXTURE_DIR = Path(__file__).resolve().parents[1]
NAMESPACE = uuid.UUID("7a0d5c7a-0000-4000-8000-0000000000e1")
CHUNKING_STRATEGY = ChunkingStrategy(max_tokens=512, overlap_tokens=50)
FIXTURE_USER_ID = uuid.UUID("10000000-0000-4000-8000-000000000001")


def _doc_id(document_key: str) -> uuid.UUID:
    return uuid.uuid5(NAMESPACE, f"doc:{document_key}")


def _chunk_id(document_key: str, chunk_index: int) -> uuid.UUID:
    return uuid.uuid5(NAMESPACE, f"chunk:{document_key}:{chunk_index}")


async def main() -> None:
    from app.config import get_settings

    settings = get_settings()
    dry_run = "--dry-run" in sys.argv
    skip_embeddings = "--skip-embeddings" in sys.argv

    fixture_manifest_path = FIXTURE_DIR / "manifest.json"
    import json

    manifest = json.loads(fixture_manifest_path.read_text())
    documents = manifest["documents"]

    parser_registry = ParserRegistry()
    chunking_registry = ChunkingRegistry()

    embedding_port = None
    if not dry_run and not skip_embeddings:
        import httpx

        embedding_port = create_embedding_adapter(settings, http_client=httpx.AsyncClient())

    if dry_run:
        print(f"DRY RUN: {len(documents)} documents to provision")
    elif embedding_port is not None:
        await embedding_port.embed_documents(["probe"])  # type: ignore[union-attr]

    session_factory = async_session_factory
    async with session_factory() as session:
        repo = SqlAlchemyKnowledgeRepository(session)
        stats: dict[str, int] = {"documents": 0, "chunks": 0}

        # Ensure the fixture user's quota can never block a 36-item eval run.
        quota_repo = SqlAlchemyQuotaRepository(session)
        existing = await quota_repo.get_quota(FIXTURE_USER_ID)
        if existing is None or (
            existing.daily_query_limit < 500 or existing.max_conversations_per_day < 200
        ):
            await quota_repo.upsert_quota(
                AIUserQuota(
                    user_id=FIXTURE_USER_ID,
                    tier=Tier.PRO,
                    daily_query_limit=1000,
                    daily_token_limit=5_000_000,
                    max_conversations_per_day=200,
                )
            )
            print("seeded fixture user quota: 1000 queries/day, 200 conversations/day")
        for doc in documents:
            document_key = doc["document_key"]
            corpus_path = (
                FIXTURE_DIR / "corpus" / doc["locale"] / f"{document_key.split(':')[-1]}.md"
            )
            if not corpus_path.exists():
                raise FileNotFoundError(f"corpus file missing: {corpus_path}")

            parser = parser_registry.get_parser("text/markdown")
            if parser is None:
                raise RuntimeError("markdown parser not registered")
            parsed = await parser.parse(
                corpus_path.read_bytes(), metadata={"filename": corpus_path.name}
            )
            chunks = await chunking_registry.chunk("structural", parsed, CHUNKING_STRATEGY)

            chunk_texts = [c.chunk_text for c in chunks]
            embeddings: list[list[float]] | None = None
            if embedding_port is not None:
                embeddings = await embedding_port.embed_documents(chunk_texts)

            knowledge_doc = KnowledgeDocument(
                id=_doc_id(document_key),
                title=doc["title"],
                source=DocumentSource(doc["source"]),
                external_id=doc["external_id"],
                effective_date=date.fromisoformat(doc["effective_date"]),
                language=Language(doc["locale"]),
                status=DocumentStatus.ACTIVE,
                version=1,
                processed_at=datetime.now(),
                metadata={"fixture": True, "fixture_version": manifest["fixture_version"]},
            )
            await repo.upsert_document(knowledge_doc)

            orm_chunks: list[DocumentChunk] = []
            for idx, chunk in enumerate(chunks):
                orm_chunks.append(
                    DocumentChunk(
                        id=_chunk_id(document_key, idx),
                        document_id=knowledge_doc.id,
                        content_type=DocumentSource(doc["source"]),
                        language=Language(doc["locale"]),
                        chunk_text=chunk.chunk_text,
                        chunk_index=chunk.provenance.chunk_index,
                        token_count=chunk.token_count,
                        embedding=embeddings[idx] if embeddings else None,
                        status=ChunkStatus.EMBEDDED if embeddings else ChunkStatus.PENDING,
                        section_heading=chunk.provenance.section_heading,
                        metadata={"fixture": True, "fixture_version": manifest["fixture_version"]},
                    )
                )
            await repo.upsert_chunks(orm_chunks)
            stats["documents"] += 1
            stats["chunks"] += len(orm_chunks)
            print(
                f"{document_key:42s} chunks={len(orm_chunks):2d} "
                f"status={'EMBEDDED' if embeddings else 'PENDING'}"
            )
        if not dry_run:
            await session.commit()
        print(f"\nprovisioned {stats['documents']} documents, {stats['chunks']} chunks")


if __name__ == "__main__":
    asyncio.run(main())
