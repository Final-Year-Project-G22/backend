from __future__ import annotations

import uuid
from abc import ABC, abstractmethod
from datetime import datetime

from core.domain.enums import DocumentSource, DocumentStatus, Language
from core.domain.models import DocumentChunk, KnowledgeDocument
from core.domain.value_objects import SearchFilters, SearchHit


class KnowledgeRepositoryPort(ABC):
    @abstractmethod
    async def upsert_document(self, document: KnowledgeDocument) -> KnowledgeDocument: ...

    @abstractmethod
    async def get_document(self, document_id: uuid.UUID) -> KnowledgeDocument | None: ...

    @abstractmethod
    async def get_document_by_external_id(
        self,
        external_id: str,
        *,
        source: DocumentSource | None = None,
    ) -> KnowledgeDocument | None: ...

    @abstractmethod
    async def list_documents(
        self,
        *,
        language: Language | None = None,
        source: DocumentSource | None = None,
        status: DocumentStatus | None = None,
        limit: int = 100,
        offset: int = 0,
    ) -> list[KnowledgeDocument]: ...

    @abstractmethod
    async def soft_delete_document(
        self, document_id: uuid.UUID, *, deleted_at: datetime
    ) -> bool: ...

    @abstractmethod
    async def upsert_chunks(self, chunks: list[DocumentChunk]) -> int: ...

    @abstractmethod
    async def get_chunks_by_document(self, document_id: uuid.UUID) -> list[DocumentChunk]: ...

    @abstractmethod
    async def delete_chunks_by_document(self, document_id: uuid.UUID) -> int: ...

    @abstractmethod
    async def complete_ingestion_atomically(
        self,
        document_id: uuid.UUID,
        *,
        status: DocumentStatus,
        chunks: list[DocumentChunk],
    ) -> tuple[KnowledgeDocument, int]: ...

    @abstractmethod
    async def search_vector(
        self,
        query_embedding: list[float],
        *,
        top_k: int,
        filters: SearchFilters | None = None,
    ) -> list[SearchHit]: ...

    @abstractmethod
    async def search_bm25(
        self,
        query: str,
        *,
        top_k: int,
        filters: SearchFilters | None = None,
    ) -> list[SearchHit]: ...


__all__ = ["KnowledgeRepositoryPort"]
