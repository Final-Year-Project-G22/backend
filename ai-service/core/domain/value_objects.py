from __future__ import annotations

import uuid
from datetime import date
from typing import Any

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator

from core.domain.enums import DocumentSource, IngestionStage, Language, Tier


class TokenUsage(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    prompt_tokens: int = Field(ge=0)
    completion_tokens: int = Field(ge=0)
    total_tokens: int = Field(ge=0)

    @model_validator(mode="after")
    def validate_total_tokens(self) -> TokenUsage:
        expected_total = self.prompt_tokens + self.completion_tokens
        if self.total_tokens != expected_total:
            msg = "total_tokens must equal prompt_tokens + completion_tokens"
            raise ValueError(msg)
        return self


class ResponseSource(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    source: DocumentSource
    document_id: uuid.UUID
    chunk_id: uuid.UUID | None = None
    title: str = Field(min_length=1, max_length=500)
    excerpt: str | None = None
    score: float | None = Field(default=None, ge=0.0)


class SearchFilters(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    language: Language | None = None
    sources: list[DocumentSource] | None = None
    effective_on: date | None = None
    only_active: bool = True
    metadata: dict[str, Any] | None = None

    @field_validator("sources")
    @classmethod
    def validate_sources(cls, value: list[DocumentSource] | None) -> list[DocumentSource] | None:
        if value is None:
            return None
        if not value:
            msg = "sources cannot be empty when provided"
            raise ValueError(msg)
        return list(dict.fromkeys(value))


class SearchHit(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    chunk_id: uuid.UUID
    document_id: uuid.UUID
    score: float = Field(ge=0.0)
    chunk_text: str = Field(min_length=1)
    chunk_index: int = Field(ge=0)
    source: DocumentSource
    language: Language
    metadata: dict[str, Any] = Field(default_factory=dict)


class UsageSnapshot(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    tier: Tier
    daily_query_count: int = Field(ge=0)
    daily_token_count: int = Field(ge=0)
    daily_conversations_started: int = Field(ge=0)
    daily_query_limit: int = Field(ge=0)
    daily_token_limit: int = Field(ge=0)
    max_conversations_per_day: int = Field(ge=0)
    total_queries_used: int = Field(ge=0)
    total_tokens_used: int = Field(ge=0)
    total_conversations: int = Field(ge=0)

    @model_validator(mode="after")
    def validate_daily_limits(self) -> UsageSnapshot:
        if self.daily_query_count > self.daily_query_limit:
            msg = "daily_query_count cannot exceed daily_query_limit"
            raise ValueError(msg)
        if self.daily_token_count > self.daily_token_limit:
            msg = "daily_token_count cannot exceed daily_token_limit"
            raise ValueError(msg)
        if self.daily_conversations_started > self.max_conversations_per_day:
            msg = "daily_conversations_started cannot exceed max_conversations_per_day"
            raise ValueError(msg)
        return self


class IngestionTransitionContext(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    event_id: str = Field(min_length=1)
    document_id: uuid.UUID
    account_id: uuid.UUID
    idempotency_key: str = Field(min_length=1)
    from_stage: IngestionStage | None = None
    to_stage: IngestionStage
    occurred_at: str = Field(min_length=1)
    retry_count: int = Field(default=0, ge=0)
    metadata: dict[str, Any] = Field(default_factory=dict)


class IngestionTransitionResult(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    context: IngestionTransitionContext
    is_terminal: bool
    status: str = Field(min_length=1)


__all__ = [
    "IngestionTransitionContext",
    "IngestionTransitionResult",
    "ResponseSource",
    "SearchFilters",
    "SearchHit",
    "TokenUsage",
    "UsageSnapshot",
]
