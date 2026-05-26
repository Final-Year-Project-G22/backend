from __future__ import annotations

import uuid

from pydantic import BaseModel, ConfigDict, Field, field_validator

from core.domain.enums import Language
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.value_objects import SearchHit, UsageSnapshot
from core.usecases.defaults import (
    DEFAULT_BM25_TOP_K,
    DEFAULT_SESSION_LIST_LIMIT,
    DEFAULT_VECTOR_TOP_K,
    MAX_PROMPT_LENGTH,
    MAX_TOP_K,
    MIN_TOP_K,
)


class CreateSessionCommand(BaseModel):
    model_config = ConfigDict(extra="forbid")

    user_id: uuid.UUID
    title: str = Field(default="New conversation", min_length=1, max_length=500)
    language: Language = Language.ENGLISH

    @field_validator("title")
    @classmethod
    def validate_title(cls, value: str) -> str:
        stripped = value.strip()
        if not stripped:
            msg = "title cannot be empty"
            raise ValueError(msg)
        return stripped


class ListSessionsQuery(BaseModel):
    model_config = ConfigDict(extra="forbid")

    user_id: uuid.UUID
    limit: int = Field(default=DEFAULT_SESSION_LIST_LIMIT, ge=1, le=200)
    offset: int = Field(default=0, ge=0)
    include_deleted: bool = False


class AskAICommand(BaseModel):
    model_config = ConfigDict(extra="forbid")

    user_id: uuid.UUID
    account_id: uuid.UUID = Field(default_factory=lambda: uuid.UUID(int=0))
    prompt: str = Field(min_length=1, max_length=MAX_PROMPT_LENGTH)
    language: Language = Language.ENGLISH
    conversation_id: uuid.UUID | None = None
    title: str | None = None
    vector_top_k: int = Field(default=DEFAULT_VECTOR_TOP_K, ge=MIN_TOP_K, le=MAX_TOP_K)
    bm25_top_k: int = Field(default=DEFAULT_BM25_TOP_K, ge=MIN_TOP_K, le=MAX_TOP_K)
    strategy: str = "simple"
    debug_mode: bool = False

    @field_validator("prompt")
    @classmethod
    def validate_prompt(cls, value: str) -> str:
        stripped = value.strip()
        if not stripped:
            msg = "prompt cannot be empty"
            raise ValueError(msg)
        return stripped


class AskAIResult(BaseModel):
    model_config = ConfigDict(extra="forbid")

    conversation: AIConversationSession
    user_message: AIChatMessage
    ai_message: AIChatMessage
    retrieved_hits: list[SearchHit] = Field(default_factory=list)
    usage: UsageSnapshot | None = None
    cache_hit: bool = False


__all__ = [
    "AskAICommand",
    "AskAIResult",
    "CreateSessionCommand",
    "ListSessionsQuery",
]
