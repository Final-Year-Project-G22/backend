from __future__ import annotations

import uuid
from datetime import UTC, date, datetime

from core.domain.enums import (
    ChunkStatus,
    DocumentSource,
    DocumentStatus,
    Language,
    MessageType,
    SessionStatus,
    Tier,
)
from core.domain.models import (
    AIChatMessage,
    AIConversationSession,
    AIUserQuota,
    DocumentChunk,
    KnowledgeDocument,
)
from core.domain.value_objects import ResponseSource, SearchHit, TokenUsage
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.mappers import (
    to_domain_chunk,
    to_domain_document,
    to_domain_message,
    to_domain_quota,
    to_domain_session,
    to_orm_chunk,
    to_orm_document,
    to_orm_message,
    to_orm_quota,
    to_orm_session,
)


def test_quota_mapper_round_trip() -> None:
    now = datetime(2026, 4, 7, 11, 0, tzinfo=UTC)
    domain_quota = AIUserQuota(
        user_id=uuid.uuid4(),
        tier=Tier.PRO,
        tier_updated_at=now,
        daily_query_count=2,
        daily_token_count=100,
        daily_conversations_started=1,
        daily_query_limit=20,
        daily_token_limit=1000,
        max_conversations_per_day=5,
        total_queries_used=50,
        total_tokens_used=5000,
        total_conversations=12,
        last_query_date=date(2026, 4, 7),
        last_query_at=now,
        last_conversation_started=now,
        created_at=now,
        updated_at=now,
    )

    orm_quota = to_orm_quota(domain_quota)
    mapped_back = to_domain_quota(orm_quota)

    assert mapped_back == domain_quota


def test_session_mapper_round_trip() -> None:
    now = datetime(2026, 4, 7, 9, 0, tzinfo=UTC)
    domain_session = AIConversationSession(
        user_id=uuid.uuid4(),
        title="Tax filing help",
        language=Language.AMHARIC,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.PRO,
        status=SessionStatus.ACTIVE,
        context_summary="Summary",
        last_message_preview="Last message",
        message_count=1,
        total_tokens_used=42,
        last_message_at=now,
        created_at=now,
        updated_at=now,
    )

    orm_session = to_orm_session(domain_session)
    mapped_back = to_domain_session(orm_session)

    assert mapped_back == domain_session


def test_message_mapper_round_trip_with_nested_values() -> None:
    now = datetime(2026, 4, 7, 10, 30, tzinfo=UTC)
    source = ResponseSource(
        source=DocumentSource.GUIDE,
        document_id=uuid.uuid4(),
        chunk_id=uuid.uuid4(),
        title="Business registration",
        excerpt="Go to city office",
        score=0.91,
    )
    search_hit = SearchHit(
        chunk_id=uuid.uuid4(),
        document_id=source.document_id,
        score=0.87,
        chunk_text="Bring your ID and forms.",
        chunk_index=0,
        source=DocumentSource.GUIDE,
        language=Language.ENGLISH,
    )
    domain_message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Please submit these documents.",
        query_language=Language.ENGLISH,
        retrieved_chunk_ids=[search_hit.chunk_id],
        context_chunks=[search_hit],
        response_sources=[source],
        token_usage=TokenUsage(prompt_tokens=10, completion_tokens=20, total_tokens=30),
        model_used="gemini-1.5-flash",
        cache_hit=True,
        message_order=2,
        created_at=now,
        updated_at=now,
    )

    orm_message = to_orm_message(domain_message)
    mapped_back = to_domain_message(orm_message)

    assert mapped_back == domain_message
    assert isinstance(orm_message.context_chunks, list)
    assert isinstance(orm_message.context_chunks[0], dict)
    assert isinstance(orm_message.response_sources, list)
    assert isinstance(orm_message.response_sources[0], dict)
    assert isinstance(orm_message.token_usage, dict)


def test_to_domain_message_uses_safe_defaults_for_optional_collections() -> None:
    now = datetime(2026, 4, 7, 8, 0, tzinfo=UTC)
    orm_message = sa_models.AIChatMessage(
        id=uuid.uuid4(),
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.USER_QUERY.value,
        user_query="How do I register my business?",
        query_language=Language.ENGLISH.value,
        query_embedding=None,
        retrieved_chunk_ids=None,
        context_chunks=None,
        llm_response=None,
        response_sources=None,
        processing_time_ms=0,
        token_usage=None,
        model_used="",
        prompt_version=1,
        trace_id=None,
        cache_hit=False,
        user_feedback=None,
        feedback_at=None,
        is_context_cleared=False,
        message_order=1,
        created_at=now,
        updated_at=now,
    )

    mapped = to_domain_message(orm_message)

    assert mapped.retrieved_chunk_ids == []
    assert mapped.response_sources == []
    assert mapped.context_chunks is None


def test_document_mapper_round_trip() -> None:
    now = datetime(2026, 4, 7, 8, 30, tzinfo=UTC)
    domain_document = KnowledgeDocument(
        title="Tax procedures",
        source=DocumentSource.GOVERNMENT,
        external_id="tax-001",
        source_url="https://example.com/tax-procedures",
        effective_date=date(2026, 1, 1),
        expiry_date=date(2026, 12, 31),
        language=Language.ENGLISH,
        status=DocumentStatus.ACTIVE,
        version=2,
        uploaded_by=uuid.uuid4(),
        processed_at=now,
        metadata={"topic": "tax"},
        created_at=now,
        updated_at=now,
    )

    orm_document = to_orm_document(domain_document)
    mapped_back = to_domain_document(orm_document)

    assert mapped_back == domain_document


def test_chunk_mapper_round_trip() -> None:
    now = datetime(2026, 4, 7, 8, 45, tzinfo=UTC)
    domain_chunk = DocumentChunk(
        document_id=uuid.uuid4(),
        content_type=DocumentSource.GUIDE,
        language=Language.AMHARIC,
        chunk_text="Bring the required forms.",
        chunk_index=0,
        token_count=5,
        embedding=[0.1, 0.2, 0.3],
        status=ChunkStatus.EMBEDDED,
        section_heading="Registration",
        metadata={"section": "registration"},
        created_at=now,
        updated_at=now,
    )

    orm_chunk = to_orm_chunk(domain_chunk)
    mapped_back = to_domain_chunk(orm_chunk)

    assert mapped_back == domain_chunk
