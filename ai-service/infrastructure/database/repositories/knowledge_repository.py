from __future__ import annotations

import uuid
from collections.abc import Mapping
from datetime import datetime
from typing import Any, cast

import numpy as np
from sqlalchemy import delete, func, literal, select
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.enums import ChunkStatus, DocumentSource, DocumentStatus, Language
from core.domain.exceptions import RepositoryError
from core.domain.models import DocumentChunk, KnowledgeDocument
from core.domain.value_objects import SearchFilters, SearchHit
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.mappers import (
    to_domain_chunk,
    to_domain_document,
    to_orm_chunk,
    to_orm_document,
)


def _normalize_query_embedding_for_pgvector(query_embedding: list[float]) -> list[float]:
    """pgvector bind processor requires a 1-D vector; batch-shaped inputs must be flattened."""
    if not query_embedding:
        return []
    arr = np.asarray(query_embedding, dtype=np.float64)
    if arr.ndim == 0:
        raise RepositoryError(
            "failed to run vector search",
            details={"reason": "query embedding must be a vector"},
        )
    if arr.ndim == 1:
        return [float(x) for x in arr.tolist()]
    expected_2d = 2
    if arr.ndim == expected_2d and arr.shape[0] == 1:
        return [float(x) for x in arr[0].tolist()]
    raise RepositoryError(
        "failed to run vector search",
        details={
            "reason": "query embedding must be 1-D",
            "shape": [int(arr.shape[i]) for i in range(arr.ndim)],
        },
    )


class SqlAlchemyKnowledgeRepository(KnowledgeRepositoryPort):
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    @property
    def session(self) -> AsyncSession:
        return self._session

    async def upsert_document(self, document: KnowledgeDocument) -> KnowledgeDocument:
        try:
            model = await self._get_document_model(document.id)
            if model is None:
                model = to_orm_document(document)
                self._session.add(model)
            else:
                self._apply_document_domain_values(model, document)
            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to upsert knowledge document",
                details={"document_id": str(document.id)},
            ) from exc

        return to_domain_document(model)

    async def get_document(self, document_id: uuid.UUID) -> KnowledgeDocument | None:
        try:
            model = await self._get_document_model(document_id)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to fetch knowledge document",
                details={"document_id": str(document_id)},
            ) from exc

        if model is None or model.deleted_at is not None:
            return None
        return to_domain_document(model)

    async def get_document_by_external_id(
        self,
        external_id: str,
        *,
        source: DocumentSource | None = None,
    ) -> KnowledgeDocument | None:
        statement = select(sa_models.KnowledgeDocument).where(
            sa_models.KnowledgeDocument.external_id == external_id,
            sa_models.KnowledgeDocument.deleted_at.is_(None),
        )
        if source is not None:
            statement = statement.where(sa_models.KnowledgeDocument.source == source.value)

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to fetch document by external id",
                details={"external_id": external_id},
            ) from exc

        model = result.scalar_one_or_none()
        if model is None:
            return None
        return to_domain_document(model)

    async def list_documents(
        self,
        *,
        language: Language | None = None,
        source: DocumentSource | None = None,
        status: DocumentStatus | None = None,
        limit: int = 100,
        offset: int = 0,
    ) -> list[KnowledgeDocument]:
        statement = (
            select(sa_models.KnowledgeDocument)
            .where(sa_models.KnowledgeDocument.deleted_at.is_(None))
            .order_by(
                sa_models.KnowledgeDocument.effective_date.desc(),
                sa_models.KnowledgeDocument.created_at.desc(),
            )
            .offset(offset)
        )
        if language is not None:
            statement = statement.where(sa_models.KnowledgeDocument.language == language.value)
        if source is not None:
            statement = statement.where(sa_models.KnowledgeDocument.source == source.value)
        if status is not None:
            statement = statement.where(sa_models.KnowledgeDocument.status == status.value)
        if limit > 0:
            statement = statement.limit(limit)

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError("failed to list knowledge documents") from exc

        models = result.scalars().all()
        return [to_domain_document(model) for model in models]

    async def soft_delete_document(self, document_id: uuid.UUID, *, deleted_at: datetime) -> bool:
        try:
            model = await self._get_document_model(document_id)
            if model is None:
                return False

            model.status = DocumentStatus.ARCHIVED.value
            model.deleted_at = deleted_at
            model.updated_at = deleted_at
            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to soft-delete knowledge document",
                details={"document_id": str(document_id)},
            ) from exc

        return True

    async def upsert_chunks(self, chunks: list[DocumentChunk]) -> int:
        if not chunks:
            return 0

        try:
            chunk_ids = [chunk.id for chunk in chunks]
            existing_result = await self._session.execute(
                select(sa_models.DocumentChunk).where(sa_models.DocumentChunk.id.in_(chunk_ids))
            )
            existing_models = {model.id: model for model in existing_result.scalars().all()}

            for chunk in chunks:
                model = existing_models.get(chunk.id)
                if model is None:
                    self._session.add(to_orm_chunk(chunk))
                else:
                    self._apply_chunk_domain_values(model, chunk)

            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError("failed to upsert document chunks") from exc

        return len(chunks)

    async def get_chunks_by_document(self, document_id: uuid.UUID) -> list[DocumentChunk]:
        statement = (
            select(sa_models.DocumentChunk)
            .where(sa_models.DocumentChunk.document_id == document_id)
            .order_by(
                sa_models.DocumentChunk.chunk_index.asc(),
                sa_models.DocumentChunk.id.asc(),
            )
        )

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError("failed to fetch document chunks") from exc

        return [to_domain_chunk(model) for model in result.scalars().all()]

    async def complete_ingestion_atomically(
        self,
        document_id: uuid.UUID,
        *,
        status: DocumentStatus,
        chunks: list[DocumentChunk],
    ) -> tuple[KnowledgeDocument, int]:
        document_result = await self._session.execute(
            select(sa_models.KnowledgeDocument).where(sa_models.KnowledgeDocument.id == document_id)
        )
        document_model = document_result.scalar_one_or_none()
        if document_model is None:
            raise RepositoryError(
                "document not found",
                details={"document_id": str(document_id)},
            )

        try:
            document_model.status = status.value
            document_model.updated_at = func.now()
            if status is DocumentStatus.ACTIVE:
                document_model.processed_at = func.now()

            existing_chunk_ids_result = await self._session.execute(
                select(sa_models.DocumentChunk.id).where(
                    sa_models.DocumentChunk.document_id == document_id
                )
            )
            existing_chunk_ids = {row[0] for row in existing_chunk_ids_result.all()}

            for chunk in chunks:
                if chunk.id in existing_chunk_ids:
                    continue
                orm_chunk = to_orm_chunk(chunk)
                orm_chunk.document_id = document_id
                self._session.add(orm_chunk)

            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError("failed to complete ingestion atomically") from exc

        # Reload document to avoid expired-attribute lazy-loading in async session
        refreshed_result = await self._session.execute(
            select(sa_models.KnowledgeDocument).where(sa_models.KnowledgeDocument.id == document_id)
        )
        document_model = refreshed_result.scalar_one()
        document = to_domain_document(document_model)
        return document, len(chunks)

    async def delete_chunks_by_document(self, document_id: uuid.UUID) -> int:
        try:
            count_result = await self._session.execute(
                select(func.count(sa_models.DocumentChunk.id)).where(
                    sa_models.DocumentChunk.document_id == document_id
                )
            )
            existing_count = count_result.scalar_one()

            await self._session.execute(
                delete(sa_models.DocumentChunk).where(
                    sa_models.DocumentChunk.document_id == document_id
                )
            )
            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to delete chunks by document",
                details={"document_id": str(document_id)},
            ) from exc

        return existing_count

    async def search_vector(
        self,
        query_embedding: list[float],
        *,
        top_k: int,
        filters: SearchFilters | None = None,
    ) -> list[SearchHit]:
        if top_k <= 0 or not query_embedding:
            return []

        vec = _normalize_query_embedding_for_pgvector(query_embedding)
        distance_expr = sa_models.DocumentChunk.embedding.op("<=>")(vec)
        score_expr = func.greatest(literal(0.0), literal(1.0) - distance_expr)

        statement = (
            select(
                sa_models.DocumentChunk.id.label("chunk_id"),
                sa_models.DocumentChunk.document_id,
                score_expr.label("score"),
                sa_models.DocumentChunk.chunk_text,
                sa_models.DocumentChunk.chunk_index,
                sa_models.KnowledgeDocument.source,
                sa_models.DocumentChunk.language,
                sa_models.DocumentChunk.metadata_.label("metadata"),
                sa_models.KnowledgeDocument.title.label("document_title"),
            )
            .join(
                sa_models.KnowledgeDocument,
                sa_models.KnowledgeDocument.id == sa_models.DocumentChunk.document_id,
            )
            .outerjoin(
                sa_models.EmbeddingProfile,
                sa_models.EmbeddingProfile.id == sa_models.DocumentChunk.embedding_profile_id,
            )
            .where(sa_models.DocumentChunk.embedding.is_not(None))
        )
        statement = self._apply_search_filters(statement, filters)
        statement = statement.order_by(distance_expr.asc()).limit(top_k)

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError("failed to run vector search") from exc

        return [
            self._row_to_search_hit(cast(Mapping[str, Any], row)) for row in result.mappings().all()
        ]

    async def search_bm25(
        self,
        query: str,
        *,
        top_k: int,
        filters: SearchFilters | None = None,
    ) -> list[SearchHit]:
        normalized_query = query.strip()
        if top_k <= 0 or not normalized_query:
            return []

        ts_query = func.plainto_tsquery("simple", normalized_query)
        rank_expr = func.ts_rank_cd(sa_models.DocumentChunk.content_tsvector, ts_query)

        statement = (
            select(
                sa_models.DocumentChunk.id.label("chunk_id"),
                sa_models.DocumentChunk.document_id,
                func.greatest(0.0, rank_expr).label("score"),
                sa_models.DocumentChunk.chunk_text,
                sa_models.DocumentChunk.chunk_index,
                sa_models.KnowledgeDocument.source,
                sa_models.DocumentChunk.language,
                sa_models.DocumentChunk.metadata_.label("metadata"),
                sa_models.KnowledgeDocument.title.label("document_title"),
            )
            .join(
                sa_models.KnowledgeDocument,
                sa_models.KnowledgeDocument.id == sa_models.DocumentChunk.document_id,
            )
            .outerjoin(
                sa_models.EmbeddingProfile,
                sa_models.EmbeddingProfile.id == sa_models.DocumentChunk.embedding_profile_id,
            )
            .where(sa_models.DocumentChunk.content_tsvector.op("@@")(ts_query))
        )
        statement = self._apply_search_filters(statement, filters)
        statement = statement.order_by(rank_expr.desc()).limit(top_k)

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError("failed to run bm25 search") from exc

        return [
            self._row_to_search_hit(cast(Mapping[str, Any], row)) for row in result.mappings().all()
        ]

    async def _get_document_model(
        self,
        document_id: uuid.UUID,
    ) -> sa_models.KnowledgeDocument | None:
        result = await self._session.execute(
            select(sa_models.KnowledgeDocument).where(sa_models.KnowledgeDocument.id == document_id)
        )
        return result.scalar_one_or_none()

    def _apply_document_domain_values(
        self,
        model: sa_models.KnowledgeDocument,
        document: KnowledgeDocument,
    ) -> None:
        model.title = document.title
        model.source = document.source.value
        model.external_id = document.external_id
        model.source_url = document.source_url
        model.effective_date = document.effective_date
        model.expiry_date = document.expiry_date
        model.language = document.language.value
        model.status = document.status.value
        model.version = document.version
        model.uploaded_by = document.uploaded_by
        model.processed_at = document.processed_at
        model.metadata_ = document.metadata
        model.created_at = document.created_at
        model.updated_at = document.updated_at
        model.deleted_at = document.deleted_at

    def _apply_chunk_domain_values(
        self,
        model: sa_models.DocumentChunk,
        chunk: DocumentChunk,
    ) -> None:
        model.document_id = chunk.document_id
        model.content_type = chunk.content_type.value
        model.language = chunk.language.value
        model.chunk_text = chunk.chunk_text
        model.chunk_index = chunk.chunk_index
        model.token_count = chunk.token_count
        model.embedding = chunk.embedding
        model.embedding_profile_id = chunk.embedding_profile_id
        model.status = chunk.status.value
        model.parent_id = chunk.parent_id
        model.section_heading = chunk.section_heading
        model.metadata_ = chunk.metadata
        model.created_at = chunk.created_at
        model.updated_at = chunk.updated_at

    def _apply_search_filters(
        self,
        statement: Any,
        filters: SearchFilters | None,
    ) -> Any:
        statement = statement.where(sa_models.KnowledgeDocument.deleted_at.is_(None))

        if filters is None:
            return statement.where(
                sa_models.KnowledgeDocument.status == DocumentStatus.ACTIVE.value,
                sa_models.DocumentChunk.status == ChunkStatus.EMBEDDED.value,
            )

        if filters.only_active:
            statement = statement.where(
                sa_models.KnowledgeDocument.status == DocumentStatus.ACTIVE.value,
                sa_models.DocumentChunk.status == ChunkStatus.EMBEDDED.value,
            )
        if filters.only_active_profile:
            statement = statement.where(
                (sa_models.EmbeddingProfile.is_active.is_(True))
                | (sa_models.EmbeddingProfile.id.is_(None))
            )
        if filters.language is not None:
            statement = statement.where(sa_models.DocumentChunk.language == filters.language.value)
        if filters.sources is not None:
            statement = statement.where(
                sa_models.KnowledgeDocument.source.in_([source.value for source in filters.sources])
            )
        if filters.effective_on is not None:
            statement = statement.where(
                sa_models.KnowledgeDocument.effective_date <= filters.effective_on,
                (
                    sa_models.KnowledgeDocument.expiry_date.is_(None)
                    | (sa_models.KnowledgeDocument.expiry_date >= filters.effective_on)
                ),
            )
        return statement

    def _row_to_search_hit(self, row: Mapping[str, Any]) -> SearchHit:
        metadata_raw = row.get("metadata")
        metadata_value: dict[str, Any]
        if isinstance(metadata_raw, dict):
            metadata_value = cast(dict[str, Any], metadata_raw)
        else:
            metadata_value = {}

        return SearchHit(
            chunk_id=row["chunk_id"],
            document_id=row["document_id"],
            score=float(row["score"]),
            chunk_text=row["chunk_text"],
            chunk_index=row["chunk_index"],
            source=DocumentSource(row["source"]),
            language=Language(row["language"]),
            metadata=dict(metadata_value),
            document_title=row.get("document_title", ""),
        )


__all__ = ["SqlAlchemyKnowledgeRepository"]
