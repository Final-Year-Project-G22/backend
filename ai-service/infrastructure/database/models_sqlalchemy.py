from __future__ import annotations

import uuid
from datetime import date, datetime
from typing import Any

from pgvector.sqlalchemy import Vector
from sqlalchemy import (
    BigInteger,
    Boolean,
    Computed,
    Date,
    DateTime,
    ForeignKey,
    Index,
    Integer,
    String,
    Text,
    func,
)
from sqlalchemy.dialects.postgresql import ARRAY, JSONB, UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from infrastructure.database.connection import Base


class KnowledgeDocument(Base):
    __tablename__ = "knowledge_documents"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    title: Mapped[str] = mapped_column(String(500), nullable=False)
    source: Mapped[str] = mapped_column(String(50), nullable=False, index=True)
    external_id: Mapped[str | None] = mapped_column(String(255), nullable=True, index=True)
    source_url: Mapped[str | None] = mapped_column(String(1000), nullable=True)
    effective_date: Mapped[date] = mapped_column(Date, nullable=False)
    expiry_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    language: Mapped[str] = mapped_column(String(5), nullable=False, index=True)
    status: Mapped[str] = mapped_column(
        String(20), nullable=False, index=True, default="processing"
    )
    version: Mapped[int] = mapped_column(Integer, nullable=False, default=1)
    uploaded_by: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), nullable=True)
    processed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    metadata_: Mapped[dict[str, Any]] = mapped_column("metadata", JSONB, default=dict)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )
    deleted_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True, index=True
    )

    chunks: Mapped[list[DocumentChunk]] = relationship(
        "DocumentChunk", back_populates="document", cascade="all, delete-orphan"
    )

    __table_args__ = (
        Index("ix_knowledge_documents_status_language", "status", "language"),
        Index("ix_knowledge_documents_external_id", "external_id"),
    )


class DocumentChunk(Base):
    __tablename__ = "document_chunks"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    document_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("knowledge_documents.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    content_type: Mapped[str] = mapped_column(String(50), nullable=False, index=True)
    language: Mapped[str] = mapped_column(String(5), nullable=False, index=True)
    chunk_text: Mapped[str] = mapped_column(Text, nullable=False)
    chunk_index: Mapped[int] = mapped_column(Integer, nullable=False)
    token_count: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    embedding: Mapped[list[float] | None] = mapped_column(Vector(1024), nullable=True)
    content_tsvector: Mapped[str | None] = mapped_column(
        Text,
        Computed("to_tsvector('simple', chunk_text)", persisted=True),
        nullable=True,
    )
    status: Mapped[str] = mapped_column(String(20), nullable=False, default="pending")
    parent_id: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), nullable=True)
    section_heading: Mapped[str | None] = mapped_column(String(500), nullable=True)
    metadata_: Mapped[dict[str, Any]] = mapped_column("metadata", JSONB, default=dict)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    document: Mapped[KnowledgeDocument] = relationship("KnowledgeDocument", back_populates="chunks")

    __table_args__ = (
        Index(
            "ix_document_chunks_embedding_hnsw",
            "embedding",
            postgresql_using="hnsw",
            postgresql_ops={"embedding": "vector_cosine_ops"},
            postgresql_with={"m": 16, "ef_construction": 200},
        ),
        Index("ix_document_chunks_content_tsvector", "content_tsvector", postgresql_using="gin"),
        Index("ix_document_chunks_status", "status"),
        Index("ix_document_chunks_content_type_language", "content_type", "language"),
    )


class AIUserQuota(Base):
    __tablename__ = "ai_user_quotas"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), nullable=False, unique=True, index=True
    )
    tier: Mapped[str] = mapped_column(String(20), nullable=False, default="basic")
    tier_updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    daily_query_count: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    daily_token_count: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    daily_conversations_started: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    daily_query_limit: Mapped[int] = mapped_column(Integer, nullable=False, default=10)
    daily_token_limit: Mapped[int] = mapped_column(BigInteger, nullable=False, default=50000)
    max_conversations_per_day: Mapped[int] = mapped_column(Integer, nullable=False, default=5)
    total_queries_used: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    total_tokens_used: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    total_conversations: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    last_query_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    last_query_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    last_conversation_started: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )


class AIConversationSession(Base):
    __tablename__ = "ai_conversation_sessions"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    title: Mapped[str] = mapped_column(String(500), nullable=False, default="New conversation")
    language: Mapped[str] = mapped_column(String(5), nullable=False, default="en")
    tier_at_start: Mapped[str] = mapped_column(String(20), nullable=False)
    current_tier: Mapped[str] = mapped_column(String(20), nullable=False)
    status: Mapped[str] = mapped_column(String(20), nullable=False, default="active")
    context_summary: Mapped[str | None] = mapped_column(Text, nullable=True)
    last_message_preview: Mapped[str] = mapped_column(String(200), nullable=False, default="")
    message_count: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    total_tokens_used: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    last_message_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )
    deleted_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True, index=True
    )

    __table_args__ = (Index("ix_ai_conversation_sessions_user_created", "user_id", "created_at"),)


class AIChatMessage(Base):
    __tablename__ = "ai_chat_messages"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    conversation_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("ai_conversation_sessions.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    message_type: Mapped[str] = mapped_column(String(20), nullable=False)
    user_query: Mapped[str | None] = mapped_column(Text, nullable=True)
    query_language: Mapped[str] = mapped_column(String(5), nullable=False, default="en")
    query_embedding: Mapped[list[float] | None] = mapped_column(Vector(1024), nullable=True)
    retrieved_chunk_ids: Mapped[list[uuid.UUID] | None] = mapped_column(
        ARRAY(UUID(as_uuid=True)), nullable=True
    )
    context_chunks: Mapped[list[dict[str, Any]] | None] = mapped_column(JSONB, nullable=True)
    llm_response: Mapped[str | None] = mapped_column(Text, nullable=True)
    response_sources: Mapped[list[dict[str, Any]] | None] = mapped_column(JSONB, nullable=True)
    processing_time_ms: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    token_usage: Mapped[dict[str, Any] | None] = mapped_column(JSONB, nullable=True)
    model_used: Mapped[str] = mapped_column(String(100), nullable=False, default="")
    prompt_version: Mapped[int] = mapped_column(Integer, nullable=False, default=1)
    trace_id: Mapped[str | None] = mapped_column(String(255), nullable=True)
    cache_hit: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    user_feedback: Mapped[int | None] = mapped_column(Integer, nullable=True)
    feedback_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    is_context_cleared: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    message_order: Mapped[int] = mapped_column(Integer, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    __table_args__ = (
        Index("ix_ai_chat_messages_conversation_order", "conversation_id", "message_order"),
        Index("ix_ai_chat_messages_created_at", "created_at"),
    )


class IngestionEventLedger(Base):
    __tablename__ = "ingestion_event_ledger"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    event_id: Mapped[str] = mapped_column(String(100), nullable=False, unique=True, index=True)
    idempotency_key: Mapped[str] = mapped_column(
        String(255), nullable=False, unique=True, index=True
    )
    account_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    document_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    occurred_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    __table_args__ = (
        Index(
            "ix_ingestion_event_ledger_document_occurred",
            "document_id",
            "occurred_at",
        ),
    )
