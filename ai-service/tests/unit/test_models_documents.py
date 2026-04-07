from __future__ import annotations

import uuid
from datetime import UTC, date, datetime

import pytest
from pydantic import ValidationError

from core.domain.enums import ChunkStatus, DocumentSource, DocumentStatus, Language
from core.domain.models import DocumentChunk, KnowledgeDocument


def test_knowledge_document_accepts_valid_payload() -> None:
    processed_at = datetime(2026, 4, 1, tzinfo=UTC)
    document = KnowledgeDocument(
        title="  Tax Filing Guide  ",
        source=DocumentSource.GUIDE,
        effective_date=date(2026, 1, 1),
        expiry_date=date(2026, 12, 31),
        language=Language.ENGLISH,
        status=DocumentStatus.ACTIVE,
        processed_at=processed_at,
    )

    assert document.title == "Tax Filing Guide"
    assert document.status is DocumentStatus.ACTIVE


def test_knowledge_document_rejects_expiry_before_effective_date() -> None:
    with pytest.raises(ValidationError):
        KnowledgeDocument(
            title="Legal update",
            source=DocumentSource.LEGAL,
            effective_date=date(2026, 3, 1),
            expiry_date=date(2026, 2, 28),
            language=Language.AMHARIC,
        )


def test_knowledge_document_requires_processed_at_when_active() -> None:
    with pytest.raises(ValidationError):
        KnowledgeDocument(
            title="Active doc",
            source=DocumentSource.GOVERNMENT,
            effective_date=date(2026, 1, 1),
            language=Language.ENGLISH,
            status=DocumentStatus.ACTIVE,
        )


def test_knowledge_document_requires_archived_status_when_deleted() -> None:
    with pytest.raises(ValidationError):
        KnowledgeDocument(
            title="Deleted doc",
            source=DocumentSource.FAQ,
            effective_date=date(2026, 1, 1),
            language=Language.ENGLISH,
            status=DocumentStatus.ACTIVE,
            processed_at=datetime(2026, 2, 1, tzinfo=UTC),
            deleted_at=datetime(2026, 2, 10, tzinfo=UTC),
        )


def test_document_chunk_accepts_valid_embedded_payload() -> None:
    chunk = DocumentChunk(
        document_id=uuid.uuid4(),
        content_type=DocumentSource.STEP,
        language=Language.ENGLISH,
        chunk_text="  register your business at the municipal office  ",
        chunk_index=0,
        token_count=8,
        status=ChunkStatus.EMBEDDED,
        embedding=[0.1, 0.2, 0.3],
    )

    assert chunk.chunk_text == "register your business at the municipal office"


def test_document_chunk_requires_embedding_when_embedded() -> None:
    with pytest.raises(ValidationError):
        DocumentChunk(
            document_id=uuid.uuid4(),
            content_type=DocumentSource.STEP,
            language=Language.ENGLISH,
            chunk_text="chunk",
            chunk_index=1,
            token_count=3,
            status=ChunkStatus.EMBEDDED,
            embedding=None,
        )


def test_document_chunk_rejects_embedding_when_not_embedded() -> None:
    with pytest.raises(ValidationError):
        DocumentChunk(
            document_id=uuid.uuid4(),
            content_type=DocumentSource.FAQ,
            language=Language.AMHARIC,
            chunk_text="chunk",
            chunk_index=2,
            token_count=2,
            status=ChunkStatus.PENDING,
            embedding=[0.4, 0.5],
        )


def test_document_chunk_rejects_non_finite_embedding_values() -> None:
    with pytest.raises(ValidationError):
        DocumentChunk(
            document_id=uuid.uuid4(),
            content_type=DocumentSource.GUIDE,
            language=Language.ENGLISH,
            chunk_text="chunk",
            chunk_index=3,
            token_count=2,
            status=ChunkStatus.EMBEDDED,
            embedding=[1.0, float("inf")],
        )
