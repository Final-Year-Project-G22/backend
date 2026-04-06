from __future__ import annotations

import math
import uuid
from datetime import UTC, date, datetime
from typing import Any

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator

from core.domain.enums import ChunkStatus, DocumentSource, DocumentStatus, Language


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


__all__ = ["DocumentChunk", "KnowledgeDocument"]
