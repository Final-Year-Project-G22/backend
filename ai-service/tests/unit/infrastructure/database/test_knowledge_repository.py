from __future__ import annotations

import uuid
from datetime import UTC, date, datetime
from typing import Any
from unittest.mock import AsyncMock, Mock

import pytest
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.enums import ChunkStatus, DocumentSource, DocumentStatus, Language
from core.domain.exceptions import RepositoryError
from core.domain.models import DocumentChunk, KnowledgeDocument
from core.domain.value_objects import SearchFilters
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.knowledge_repository import SqlAlchemyKnowledgeRepository
from infrastructure.database.repositories.mappers import to_orm_chunk, to_orm_document


class _ScalarResult:
    def __init__(
        self,
        value: sa_models.KnowledgeDocument | sa_models.DocumentChunk | int | None,
    ) -> None:
        self._value = value

    def scalar_one_or_none(self) -> sa_models.KnowledgeDocument | sa_models.DocumentChunk | None:
        return self._value

    def scalar_one(self) -> int:
        assert isinstance(self._value, int)
        return self._value


class _ScalarsResult:
    def __init__(self, values: list[sa_models.KnowledgeDocument | sa_models.DocumentChunk]) -> None:
        self._values = values

    def all(self) -> list[sa_models.KnowledgeDocument | sa_models.DocumentChunk]:
        return self._values


class _ListResult:
    def __init__(self, values: list[sa_models.KnowledgeDocument | sa_models.DocumentChunk]) -> None:
        self._values = values

    def scalars(self) -> _ScalarsResult:
        return _ScalarsResult(self._values)


class _DeleteResult:
    def __init__(self, rowcount: int) -> None:
        self.rowcount = rowcount


class _RowsResult:
    def __init__(self, rows: list[dict[str, Any]]) -> None:
        self._rows = rows

    def mappings(self) -> _RowsResult:
        return self

    def all(self) -> list[dict[str, Any]]:
        return self._rows


def _build_session() -> AsyncMock:
    session = AsyncMock(spec=AsyncSession)
    session.add = Mock()
    return session


def _build_document(*, document_id: uuid.UUID | None = None) -> KnowledgeDocument:
    now = datetime(2026, 4, 7, 9, 0, tzinfo=UTC)
    return KnowledgeDocument(
        id=document_id or uuid.uuid4(),
        title="Business registration guide",
        source=DocumentSource.GUIDE,
        external_id="guide-001",
        effective_date=date(2026, 1, 1),
        language=Language.ENGLISH,
        status=DocumentStatus.ACTIVE,
        processed_at=now,
        created_at=now,
        updated_at=now,
    )


def _build_chunk(*, chunk_id: uuid.UUID | None = None, document_id: uuid.UUID) -> DocumentChunk:
    now = datetime(2026, 4, 7, 9, 5, tzinfo=UTC)
    return DocumentChunk(
        id=chunk_id or uuid.uuid4(),
        document_id=document_id,
        content_type=DocumentSource.GUIDE,
        language=Language.ENGLISH,
        chunk_text="Visit the city office to register your business.",
        chunk_index=0,
        token_count=10,
        embedding=[0.2, 0.4, 0.6],
        status=ChunkStatus.EMBEDDED,
        created_at=now,
        updated_at=now,
    )


@pytest.mark.asyncio
async def test_upsert_document_creates_when_missing() -> None:
    session = _build_session()
    session.execute.return_value = _ScalarResult(None)
    repo = SqlAlchemyKnowledgeRepository(session)

    document = _build_document()
    result = await repo.upsert_document(document)

    assert result == document
    session.add.assert_called_once()
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_get_document_returns_none_when_soft_deleted() -> None:
    deleted_at = datetime(2026, 4, 8, 10, 0, tzinfo=UTC)
    soft_deleted = _build_document().model_copy(
        update={"status": DocumentStatus.ARCHIVED, "deleted_at": deleted_at}
    )
    orm_document = to_orm_document(soft_deleted)

    session = _build_session()
    session.execute.return_value = _ScalarResult(orm_document)
    repo = SqlAlchemyKnowledgeRepository(session)

    result = await repo.get_document(soft_deleted.id)

    assert result is None


@pytest.mark.asyncio
async def test_get_document_by_external_id_uses_source_filter() -> None:
    document = _build_document()
    orm_document = to_orm_document(document)

    session = _build_session()
    session.execute.return_value = _ScalarResult(orm_document)
    repo = SqlAlchemyKnowledgeRepository(session)

    result = await repo.get_document_by_external_id(
        document.external_id or "",
        source=DocumentSource.GUIDE,
    )

    assert result == document
    statement = session.execute.await_args.args[0]
    assert "knowledge_documents.external_id" in str(statement)
    assert "knowledge_documents.source" in str(statement)


@pytest.mark.asyncio
async def test_list_documents_applies_optional_filters() -> None:
    document = _build_document()
    orm_document = to_orm_document(document)

    session = _build_session()
    session.execute.return_value = _ListResult([orm_document])
    repo = SqlAlchemyKnowledgeRepository(session)

    results = await repo.list_documents(
        language=Language.ENGLISH,
        source=DocumentSource.GUIDE,
        status=DocumentStatus.ACTIVE,
        limit=10,
        offset=0,
    )

    assert results == [document]
    statement = session.execute.await_args.args[0]
    statement_str = str(statement)
    assert "knowledge_documents.language" in statement_str
    assert "knowledge_documents.source" in statement_str
    assert "knowledge_documents.status" in statement_str


@pytest.mark.asyncio
async def test_soft_delete_document_marks_archived() -> None:
    document = _build_document()
    orm_document = to_orm_document(document)

    session = _build_session()
    session.execute.return_value = _ScalarResult(orm_document)
    repo = SqlAlchemyKnowledgeRepository(session)

    deleted_at = datetime(2026, 4, 8, 10, 10, tzinfo=UTC)
    deleted = await repo.soft_delete_document(document.id, deleted_at=deleted_at)

    assert deleted is True
    assert orm_document.status == DocumentStatus.ARCHIVED.value
    assert orm_document.deleted_at == deleted_at


@pytest.mark.asyncio
async def test_get_chunks_by_document_returns_domain_chunks() -> None:
    document = _build_document()
    chunk = _build_chunk(document_id=document.id)
    orm_chunk = to_orm_chunk(chunk)

    session = _build_session()
    session.execute.return_value = _ListResult([orm_chunk])
    repo = SqlAlchemyKnowledgeRepository(session)

    results = await repo.get_chunks_by_document(document.id)

    assert results == [chunk]
    statement = session.execute.await_args.args[0]
    assert "document_chunks.chunk_index" in str(statement)


@pytest.mark.asyncio
async def test_upsert_chunks_creates_and_updates_models() -> None:
    expected_affected = 2

    document = _build_document()
    existing_chunk = _build_chunk(document_id=document.id)
    new_chunk = _build_chunk(document_id=document.id)
    existing_model = to_orm_chunk(existing_chunk)
    existing_model.chunk_text = "Old text"

    session = _build_session()
    session.execute.return_value = _ListResult([existing_model])
    repo = SqlAlchemyKnowledgeRepository(session)

    affected = await repo.upsert_chunks([existing_chunk, new_chunk])

    assert affected == expected_affected
    assert existing_model.chunk_text == existing_chunk.chunk_text
    session.add.assert_called_once()
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_delete_chunks_by_document_returns_deleted_row_count() -> None:
    expected_rowcount = 3

    session = _build_session()
    session.execute.side_effect = [
        _ScalarResult(expected_rowcount),
        _DeleteResult(expected_rowcount),
    ]
    repo = SqlAlchemyKnowledgeRepository(session)

    deleted = await repo.delete_chunks_by_document(uuid.uuid4())

    assert deleted == expected_rowcount
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_search_vector_maps_rows_to_search_hits() -> None:
    document_id = uuid.uuid4()
    chunk_id = uuid.uuid4()
    row = {
        "chunk_id": chunk_id,
        "document_id": document_id,
        "score": 0.82,
        "chunk_text": "Bring ID documents.",
        "chunk_index": 0,
        "source": DocumentSource.GUIDE.value,
        "language": Language.ENGLISH.value,
        "metadata": {"section": "requirements"},
    }

    session = _build_session()
    session.execute.return_value = _RowsResult([row])
    repo = SqlAlchemyKnowledgeRepository(session)

    hits = await repo.search_vector([0.1, 0.2, 0.3], top_k=5)

    assert len(hits) == 1
    assert hits[0].chunk_id == chunk_id
    assert hits[0].score == row["score"]


@pytest.mark.asyncio
async def test_search_bm25_applies_filters_and_maps_rows() -> None:
    row = {
        "chunk_id": uuid.uuid4(),
        "document_id": uuid.uuid4(),
        "score": 0.65,
        "chunk_text": "Apply for a trade license.",
        "chunk_index": 1,
        "source": DocumentSource.GOVERNMENT.value,
        "language": Language.AMHARIC.value,
        "metadata": {},
    }

    session = _build_session()
    session.execute.return_value = _RowsResult([row])
    repo = SqlAlchemyKnowledgeRepository(session)

    filters = SearchFilters(
        language=Language.AMHARIC,
        sources=[DocumentSource.GOVERNMENT],
        effective_on=date(2026, 4, 7),
    )
    hits = await repo.search_bm25("trade license", top_k=3, filters=filters)

    assert len(hits) == 1
    assert hits[0].source is DocumentSource.GOVERNMENT
    assert hits[0].language is Language.AMHARIC

    statement = session.execute.await_args.args[0]
    assert "knowledge_documents.source" in str(statement)


@pytest.mark.asyncio
async def test_search_returns_empty_for_invalid_query_inputs() -> None:
    session = _build_session()
    repo = SqlAlchemyKnowledgeRepository(session)

    vector_hits = await repo.search_vector([], top_k=3)
    bm25_hits = await repo.search_bm25("   ", top_k=3)

    assert vector_hits == []
    assert bm25_hits == []
    session.execute.assert_not_awaited()


@pytest.mark.asyncio
async def test_get_document_wraps_sqlalchemy_errors() -> None:
    session = _build_session()
    session.execute.side_effect = SQLAlchemyError("boom")
    repo = SqlAlchemyKnowledgeRepository(session)

    with pytest.raises(RepositoryError, match="failed to fetch knowledge document"):
        await repo.get_document(uuid.uuid4())
