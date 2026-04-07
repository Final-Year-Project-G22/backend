from __future__ import annotations

import math
import uuid
from datetime import UTC, date, datetime
from typing import Any

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator

from core.domain.enums import (
    ChunkStatus,
    DocumentSource,
    DocumentStatus,
    Language,
    MessageType,
    SessionStatus,
    Tier,
)
from core.domain.value_objects import ResponseSource, SearchHit, TokenUsage


def _utc_now() -> datetime:
    return datetime.now(UTC)


class KnowledgeDocument(BaseModel):
    model_config = ConfigDict(extra="forbid")

    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    title: str = Field(min_length=1, max_length=500)
    source: DocumentSource
    external_id: str | None = Field(default=None, min_length=1, max_length=255)
    source_url: str | None = Field(default=None, min_length=1, max_length=1000)
    effective_date: date
    expiry_date: date | None = None
    language: Language
    status: DocumentStatus = DocumentStatus.PROCESSING
    version: int = Field(default=1, ge=1)
    uploaded_by: uuid.UUID | None = None
    processed_at: datetime | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)
    created_at: datetime = Field(default_factory=_utc_now)
    updated_at: datetime = Field(default_factory=_utc_now)
    deleted_at: datetime | None = None

    @field_validator("title")
    @classmethod
    def validate_title(cls, value: str) -> str:
        stripped = value.strip()
        if not stripped:
            msg = "title cannot be empty"
            raise ValueError(msg)
        return stripped

    @model_validator(mode="after")
    def validate_temporal_state(self) -> KnowledgeDocument:
        if self.expiry_date is not None and self.expiry_date < self.effective_date:
            msg = "expiry_date must be on or after effective_date"
            raise ValueError(msg)

        if self.status is DocumentStatus.ACTIVE and self.processed_at is None:
            msg = "processed_at is required when status is active"
            raise ValueError(msg)

        if self.deleted_at is not None and self.status is not DocumentStatus.ARCHIVED:
            msg = "status must be archived when deleted_at is set"
            raise ValueError(msg)

        return self


class DocumentChunk(BaseModel):
    model_config = ConfigDict(extra="forbid")

    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    document_id: uuid.UUID
    content_type: DocumentSource
    language: Language
    chunk_text: str = Field(min_length=1)
    chunk_index: int = Field(ge=0)
    token_count: int = Field(ge=0)
    embedding: list[float] | None = None
    status: ChunkStatus = ChunkStatus.PENDING
    parent_id: uuid.UUID | None = None
    section_heading: str | None = Field(default=None, max_length=500)
    metadata: dict[str, Any] = Field(default_factory=dict)
    created_at: datetime = Field(default_factory=_utc_now)
    updated_at: datetime = Field(default_factory=_utc_now)

    @field_validator("chunk_text")
    @classmethod
    def validate_chunk_text(cls, value: str) -> str:
        stripped = value.strip()
        if not stripped:
            msg = "chunk_text cannot be empty"
            raise ValueError(msg)
        return stripped

    @field_validator("embedding")
    @classmethod
    def validate_embedding_values(cls, value: list[float] | None) -> list[float] | None:
        if value is None:
            return None
        if not value:
            msg = "embedding cannot be empty"
            raise ValueError(msg)
        if not all(math.isfinite(item) for item in value):
            msg = "embedding values must be finite numbers"
            raise ValueError(msg)
        return value

    @model_validator(mode="after")
    def validate_embedding_state(self) -> DocumentChunk:
        if self.status is ChunkStatus.EMBEDDED and self.embedding is None:
            msg = "embedding is required when status is embedded"
            raise ValueError(msg)
        if self.status is not ChunkStatus.EMBEDDED and self.embedding is not None:
            msg = "embedding must be null unless status is embedded"
            raise ValueError(msg)
        return self


class AIUserQuota(BaseModel):
    model_config = ConfigDict(extra="forbid")

    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    user_id: uuid.UUID
    tier: Tier = Tier.BASIC
    tier_updated_at: datetime = Field(default_factory=_utc_now)

    daily_query_count: int = Field(default=0, ge=0)
    daily_token_count: int = Field(default=0, ge=0)
    daily_conversations_started: int = Field(default=0, ge=0)

    daily_query_limit: int = Field(default=10, ge=0)
    daily_token_limit: int = Field(default=50000, ge=0)
    max_conversations_per_day: int = Field(default=5, ge=0)

    total_queries_used: int = Field(default=0, ge=0)
    total_tokens_used: int = Field(default=0, ge=0)
    total_conversations: int = Field(default=0, ge=0)

    last_query_date: date | None = None
    last_query_at: datetime | None = None
    last_conversation_started: datetime | None = None
    created_at: datetime = Field(default_factory=_utc_now)
    updated_at: datetime = Field(default_factory=_utc_now)

    @model_validator(mode="after")
    def validate_daily_limits(self) -> AIUserQuota:
        if self.daily_query_count > self.daily_query_limit:
            msg = "daily_query_count cannot exceed daily_query_limit"
            raise ValueError(msg)
        if self.daily_token_count > self.daily_token_limit:
            msg = "daily_token_count cannot exceed daily_token_limit"
            raise ValueError(msg)
        if self.daily_conversations_started > self.max_conversations_per_day:
            msg = "daily_conversations_started cannot exceed max_conversations_per_day"
            raise ValueError(msg)
        if (
            self.last_query_at is not None
            and self.last_query_date is not None
            and self.last_query_at.date() != self.last_query_date
        ):
            msg = "last_query_date must match last_query_at.date()"
            raise ValueError(msg)
        return self


class AIConversationSession(BaseModel):
    model_config = ConfigDict(extra="forbid")

    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    user_id: uuid.UUID
    title: str = Field(default="New conversation", min_length=1, max_length=500)
    language: Language = Language.ENGLISH
    tier_at_start: Tier
    current_tier: Tier
    status: SessionStatus = SessionStatus.ACTIVE
    context_summary: str | None = None
    last_message_preview: str = Field(default="", max_length=200)
    message_count: int = Field(default=0, ge=0)
    total_tokens_used: int = Field(default=0, ge=0)
    last_message_at: datetime | None = None
    created_at: datetime = Field(default_factory=_utc_now)
    updated_at: datetime = Field(default_factory=_utc_now)
    deleted_at: datetime | None = None

    @field_validator("title")
    @classmethod
    def validate_title(cls, value: str) -> str:
        stripped = value.strip()
        if not stripped:
            msg = "title cannot be empty"
            raise ValueError(msg)
        return stripped

    @model_validator(mode="after")
    def validate_session_state(self) -> AIConversationSession:
        if self.deleted_at is not None and self.status is not SessionStatus.ARCHIVED:
            msg = "status must be archived when deleted_at is set"
            raise ValueError(msg)
        if self.message_count > 0 and self.last_message_at is None:
            msg = "last_message_at is required when message_count is greater than 0"
            raise ValueError(msg)
        return self


class AIChatMessage(BaseModel):
    model_config = ConfigDict(extra="forbid")

    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    user_id: uuid.UUID
    conversation_id: uuid.UUID
    message_type: MessageType

    user_query: str | None = None
    query_language: Language = Language.ENGLISH
    query_embedding: list[float] | None = None
    retrieved_chunk_ids: list[uuid.UUID] = Field(default_factory=list)
    context_chunks: list[SearchHit] | None = None

    llm_response: str | None = None
    response_sources: list[ResponseSource] = Field(default_factory=list)
    processing_time_ms: int = Field(default=0, ge=0)
    token_usage: TokenUsage | None = None
    model_used: str = Field(default="", max_length=100)
    prompt_version: int = Field(default=1, ge=1)

    trace_id: str | None = Field(default=None, min_length=1, max_length=255)
    cache_hit: bool = False

    user_feedback: int | None = Field(default=None, ge=-1, le=5)
    feedback_at: datetime | None = None
    is_context_cleared: bool = False
    message_order: int = Field(ge=1)
    created_at: datetime = Field(default_factory=_utc_now)
    updated_at: datetime = Field(default_factory=_utc_now)

    @field_validator("user_query")
    @classmethod
    def validate_user_query(cls, value: str | None) -> str | None:
        if value is None:
            return None
        stripped = value.strip()
        if not stripped:
            msg = "user_query cannot be empty when provided"
            raise ValueError(msg)
        return stripped

    @field_validator("llm_response")
    @classmethod
    def validate_llm_response(cls, value: str | None) -> str | None:
        if value is None:
            return None
        stripped = value.strip()
        if not stripped:
            msg = "llm_response cannot be empty when provided"
            raise ValueError(msg)
        return stripped

    @field_validator("query_embedding")
    @classmethod
    def validate_query_embedding(cls, value: list[float] | None) -> list[float] | None:
        if value is None:
            return None
        if not value:
            msg = "query_embedding cannot be empty"
            raise ValueError(msg)
        if not all(math.isfinite(item) for item in value):
            msg = "query_embedding values must be finite numbers"
            raise ValueError(msg)
        return value

    @model_validator(mode="after")
    def validate_message_semantics(self) -> AIChatMessage:
        if self.message_type is MessageType.USER_QUERY:
            if self.user_query is None:
                msg = "user_query is required for user_query message type"
                raise ValueError(msg)
            if self.llm_response is not None:
                msg = "llm_response must be null for user_query message type"
                raise ValueError(msg)

        if self.message_type is MessageType.AI_RESPONSE and self.llm_response is None:
            msg = "llm_response is required for ai_response message type"
            raise ValueError(msg)

        if (self.user_feedback is None) != (self.feedback_at is None):
            msg = "user_feedback and feedback_at must either both be set or both be null"
            raise ValueError(msg)

        return self


__all__ = [
    "AIChatMessage",
    "AIConversationSession",
    "AIUserQuota",
    "DocumentChunk",
    "KnowledgeDocument",
]
