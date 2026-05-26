from __future__ import annotations

from typing import Any

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
    ToolCallRecord,
)
from core.domain.value_objects import ResponseSource, SearchHit, TokenUsage
from infrastructure.database import models_sqlalchemy as sa_models


def to_domain_quota(model: sa_models.AIUserQuota) -> AIUserQuota:
    return AIUserQuota(
        id=model.id,
        user_id=model.user_id,
        tier=Tier(model.tier),
        tier_updated_at=model.tier_updated_at,
        daily_query_count=model.daily_query_count,
        daily_token_count=model.daily_token_count,
        daily_conversations_started=model.daily_conversations_started,
        daily_query_limit=model.daily_query_limit,
        daily_token_limit=model.daily_token_limit,
        max_conversations_per_day=model.max_conversations_per_day,
        total_queries_used=model.total_queries_used,
        total_tokens_used=model.total_tokens_used,
        total_conversations=model.total_conversations,
        last_query_date=model.last_query_date,
        last_query_at=model.last_query_at,
        last_conversation_started=model.last_conversation_started,
        created_at=model.created_at,
        updated_at=model.updated_at,
    )


def to_orm_quota(domain: AIUserQuota) -> sa_models.AIUserQuota:
    return sa_models.AIUserQuota(
        id=domain.id,
        user_id=domain.user_id,
        tier=domain.tier.value,
        tier_updated_at=domain.tier_updated_at,
        daily_query_count=domain.daily_query_count,
        daily_token_count=domain.daily_token_count,
        daily_conversations_started=domain.daily_conversations_started,
        daily_query_limit=domain.daily_query_limit,
        daily_token_limit=domain.daily_token_limit,
        max_conversations_per_day=domain.max_conversations_per_day,
        total_queries_used=domain.total_queries_used,
        total_tokens_used=domain.total_tokens_used,
        total_conversations=domain.total_conversations,
        last_query_date=domain.last_query_date,
        last_query_at=domain.last_query_at,
        last_conversation_started=domain.last_conversation_started,
        created_at=domain.created_at,
        updated_at=domain.updated_at,
    )


def to_domain_session(model: sa_models.AIConversationSession) -> AIConversationSession:
    return AIConversationSession(
        id=model.id,
        user_id=model.user_id,
        title=model.title,
        language=Language(model.language),
        tier_at_start=Tier(model.tier_at_start),
        current_tier=Tier(model.current_tier),
        status=SessionStatus(model.status),
        context_summary=model.context_summary,
        last_message_preview=model.last_message_preview,
        message_count=model.message_count,
        total_tokens_used=model.total_tokens_used,
        last_message_at=model.last_message_at,
        created_at=model.created_at,
        updated_at=model.updated_at,
        deleted_at=model.deleted_at,
    )


def to_orm_session(domain: AIConversationSession) -> sa_models.AIConversationSession:
    return sa_models.AIConversationSession(
        id=domain.id,
        user_id=domain.user_id,
        title=domain.title,
        language=domain.language.value,
        tier_at_start=domain.tier_at_start.value,
        current_tier=domain.current_tier.value,
        status=domain.status.value,
        context_summary=domain.context_summary,
        last_message_preview=domain.last_message_preview,
        message_count=domain.message_count,
        total_tokens_used=domain.total_tokens_used,
        last_message_at=domain.last_message_at,
        created_at=domain.created_at,
        updated_at=domain.updated_at,
        deleted_at=domain.deleted_at,
    )


def to_domain_message(model: sa_models.AIChatMessage) -> AIChatMessage:
    return AIChatMessage(
        id=model.id,
        user_id=model.user_id,
        conversation_id=model.conversation_id,
        message_type=MessageType(model.message_type),
        user_query=model.user_query,
        query_language=Language(model.query_language),
        query_embedding=model.query_embedding,
        retrieved_chunk_ids=model.retrieved_chunk_ids or [],
        context_chunks=_deserialize_search_hits(model.context_chunks),
        llm_response=model.llm_response,
        response_sources=_deserialize_response_sources(model.response_sources),
        processing_time_ms=model.processing_time_ms,
        token_usage=_deserialize_token_usage(model.token_usage),
        model_used=model.model_used,
        prompt_version=model.prompt_version,
        tool_calls=_deserialize_tool_calls(model.tool_calls),
        agent_strategy=model.agent_strategy or "simple",
        trace_id=model.trace_id,
        cache_hit=model.cache_hit,
        user_feedback=model.user_feedback,
        feedback_at=model.feedback_at,
        is_context_cleared=model.is_context_cleared,
        message_order=model.message_order,
        created_at=model.created_at,
        updated_at=model.updated_at,
    )


def to_orm_message(domain: AIChatMessage) -> sa_models.AIChatMessage:
    return sa_models.AIChatMessage(
        id=domain.id,
        user_id=domain.user_id,
        conversation_id=domain.conversation_id,
        message_type=domain.message_type.value,
        user_query=domain.user_query,
        query_language=domain.query_language.value,
        query_embedding=domain.query_embedding,
        retrieved_chunk_ids=domain.retrieved_chunk_ids,
        context_chunks=serialize_search_hits(domain.context_chunks),
        llm_response=domain.llm_response,
        response_sources=serialize_response_sources(domain.response_sources),
        processing_time_ms=domain.processing_time_ms,
        token_usage=serialize_token_usage(domain.token_usage),
        model_used=domain.model_used,
        prompt_version=domain.prompt_version,
        tool_calls=serialize_tool_calls(domain.tool_calls),
        agent_strategy=domain.agent_strategy,
        trace_id=domain.trace_id,
        cache_hit=domain.cache_hit,
        user_feedback=domain.user_feedback,
        feedback_at=domain.feedback_at,
        is_context_cleared=domain.is_context_cleared,
        message_order=domain.message_order,
        created_at=domain.created_at,
        updated_at=domain.updated_at,
    )


def to_domain_document(model: sa_models.KnowledgeDocument) -> KnowledgeDocument:
    return KnowledgeDocument(
        id=model.id,
        title=model.title,
        source=DocumentSource(model.source),
        external_id=model.external_id,
        source_url=model.source_url,
        effective_date=model.effective_date,
        expiry_date=model.expiry_date,
        language=Language(model.language),
        status=DocumentStatus(model.status),
        version=model.version,
        uploaded_by=model.uploaded_by,
        processed_at=model.processed_at,
        metadata=model.metadata_ or {},
        created_at=model.created_at,
        updated_at=model.updated_at,
        deleted_at=model.deleted_at,
    )


def to_orm_document(domain: KnowledgeDocument) -> sa_models.KnowledgeDocument:
    return sa_models.KnowledgeDocument(
        id=domain.id,
        title=domain.title,
        source=domain.source.value,
        external_id=domain.external_id,
        source_url=domain.source_url,
        effective_date=domain.effective_date,
        expiry_date=domain.expiry_date,
        language=domain.language.value,
        status=domain.status.value,
        version=domain.version,
        uploaded_by=domain.uploaded_by,
        processed_at=domain.processed_at,
        metadata_=domain.metadata,
        created_at=domain.created_at,
        updated_at=domain.updated_at,
        deleted_at=domain.deleted_at,
    )


def to_domain_chunk(model: sa_models.DocumentChunk) -> DocumentChunk:
    return DocumentChunk(
        id=model.id,
        document_id=model.document_id,
        content_type=DocumentSource(model.content_type),
        language=Language(model.language),
        chunk_text=model.chunk_text,
        chunk_index=model.chunk_index,
        token_count=model.token_count,
        embedding=model.embedding,
        embedding_profile_id=model.embedding_profile_id,
        status=ChunkStatus(model.status),
        parent_id=model.parent_id,
        section_heading=model.section_heading,
        metadata=model.metadata_ or {},
        created_at=model.created_at,
        updated_at=model.updated_at,
    )


def to_orm_chunk(domain: DocumentChunk) -> sa_models.DocumentChunk:
    return sa_models.DocumentChunk(
        id=domain.id,
        document_id=domain.document_id,
        content_type=domain.content_type.value,
        language=domain.language.value,
        chunk_text=domain.chunk_text,
        chunk_index=domain.chunk_index,
        token_count=domain.token_count,
        embedding=domain.embedding,
        embedding_profile_id=domain.embedding_profile_id,
        status=domain.status.value,
        parent_id=domain.parent_id,
        section_heading=domain.section_heading,
        metadata_=domain.metadata,
        created_at=domain.created_at,
        updated_at=domain.updated_at,
    )


def serialize_token_usage(token_usage: TokenUsage | None) -> dict[str, Any] | None:
    if token_usage is None:
        return None
    return token_usage.model_dump(mode="python")


def _deserialize_token_usage(raw: dict[str, Any] | None) -> TokenUsage | None:
    if raw is None:
        return None
    return TokenUsage.model_validate(raw)


def serialize_response_sources(
    sources: list[ResponseSource],
) -> list[dict[str, Any]]:
    return [source.model_dump(mode="json") for source in sources]


def _deserialize_response_sources(raw: list[dict[str, Any]] | None) -> list[ResponseSource]:
    if raw is None:
        return []
    return [ResponseSource.model_validate(item) for item in raw]


def serialize_search_hits(hits: list[SearchHit] | None) -> list[dict[str, Any]] | None:
    if hits is None:
        return None
    return [hit.model_dump(mode="json") for hit in hits]


def _deserialize_search_hits(raw: list[dict[str, Any]] | None) -> list[SearchHit] | None:
    if raw is None:
        return None
    return [SearchHit.model_validate(item) for item in raw]


def serialize_tool_calls(tool_calls: list[ToolCallRecord] | None) -> list[dict[str, Any]] | None:
    if tool_calls is None:
        return None
    return [tc.model_dump(mode="python") for tc in tool_calls]


def _deserialize_tool_calls(raw: list[dict[str, Any]] | None) -> list[ToolCallRecord] | None:
    if raw is None:
        return None
    return [ToolCallRecord.model_validate(item) for item in raw]


__all__ = [
    "to_domain_chunk",
    "to_domain_document",
    "to_domain_message",
    "to_domain_quota",
    "to_domain_session",
    "to_orm_chunk",
    "to_orm_document",
    "to_orm_message",
    "to_orm_quota",
    "to_orm_session",
]
