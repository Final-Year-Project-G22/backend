"""create knowledge tables

Revision ID: 001
Revises:
Create Date: 2026-04-05
"""

from collections.abc import Sequence

import pgvector  # type: ignore[reportUnusedImport]
import pgvector.alembic  # type: ignore[import-not-found,reportUnusedImport]
import sqlalchemy as sa
from pgvector.sqlalchemy import Vector
from sqlalchemy.dialects import postgresql

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "001"
down_revision: str | None = None
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS vector")

    op.create_table(
        "knowledge_documents",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("title", sa.String(500), nullable=False),
        sa.Column("source", sa.String(50), nullable=False),
        sa.Column("external_id", sa.String(255), nullable=True),
        sa.Column("source_url", sa.String(1000), nullable=True),
        sa.Column("effective_date", sa.Date, nullable=False),
        sa.Column("expiry_date", sa.Date, nullable=True),
        sa.Column("language", sa.String(5), nullable=False),
        sa.Column("status", sa.String(20), nullable=False, server_default="processing"),
        sa.Column("version", sa.Integer, nullable=False, server_default="1"),
        sa.Column("uploaded_by", postgresql.UUID(as_uuid=True), nullable=True),
        sa.Column("processed_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("metadata", postgresql.JSONB, server_default="{}"),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            onupdate=sa.func.now(),
        ),
        sa.Column("deleted_at", sa.DateTime(timezone=True), nullable=True),
    )

    op.create_index("ix_knowledge_documents_source", "knowledge_documents", ["source"])
    op.create_index("ix_knowledge_documents_external_id", "knowledge_documents", ["external_id"])
    op.create_index("ix_knowledge_documents_language", "knowledge_documents", ["language"])
    op.create_index("ix_knowledge_documents_status", "knowledge_documents", ["status"])
    op.create_index(
        "ix_knowledge_documents_status_language",
        "knowledge_documents",
        ["status", "language"],
    )
    op.create_index("ix_knowledge_documents_deleted_at", "knowledge_documents", ["deleted_at"])

    op.create_table(
        "document_chunks",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "document_id",
            postgresql.UUID(as_uuid=True),
            sa.ForeignKey("knowledge_documents.id", ondelete="CASCADE"),
            nullable=False,
        ),
        sa.Column("content_type", sa.String(50), nullable=False),
        sa.Column("language", sa.String(5), nullable=False),
        sa.Column("chunk_text", sa.Text, nullable=False),
        sa.Column("chunk_index", sa.Integer, nullable=False),
        sa.Column("token_count", sa.Integer, nullable=False, server_default="0"),
        sa.Column("embedding", Vector(1024), nullable=True),  # type: ignore[reportUnknownArgumentType]
        sa.Column("status", sa.String(20), nullable=False, server_default="pending"),
        sa.Column("parent_id", postgresql.UUID(as_uuid=True), nullable=True),
        sa.Column("section_heading", sa.String(500), nullable=True),
        sa.Column("metadata", postgresql.JSONB, server_default="{}"),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            onupdate=sa.func.now(),
        ),
    )

    op.create_index("ix_document_chunks_document_id", "document_chunks", ["document_id"])
    op.create_index("ix_document_chunks_content_type", "document_chunks", ["content_type"])
    op.create_index("ix_document_chunks_language", "document_chunks", ["language"])
    op.create_index("ix_document_chunks_status", "document_chunks", ["status"])
    op.create_index(
        "ix_document_chunks_content_type_language",
        "document_chunks",
        ["content_type", "language"],
    )
    op.execute(
        """
        CREATE INDEX ix_document_chunks_embedding_hnsw
        ON document_chunks
        USING hnsw (embedding vector_cosine_ops)
        WITH (m = 16, ef_construction = 200)
        """
    )

    # Add generated column for full-text search
    op.execute(
        """
        ALTER TABLE document_chunks
        ADD COLUMN content_tsvector tsvector
        GENERATED ALWAYS AS (to_tsvector('simple', chunk_text)) STORED
        """
    )

    op.execute(
        """
        CREATE INDEX ix_document_chunks_content_tsvector
        ON document_chunks
        USING gin (content_tsvector)
        """
    )

    op.create_table(
        "ai_user_quotas",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("tier", sa.String(20), nullable=False, server_default="basic"),
        sa.Column("tier_updated_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column("daily_query_count", sa.Integer, nullable=False, server_default="0"),
        sa.Column("daily_token_count", sa.Integer, nullable=False, server_default="0"),
        sa.Column("daily_conversations_started", sa.Integer, nullable=False, server_default="0"),
        sa.Column("daily_query_limit", sa.Integer, nullable=False, server_default="10"),
        sa.Column("daily_token_limit", sa.BigInteger, nullable=False, server_default="50000"),
        sa.Column("max_conversations_per_day", sa.Integer, nullable=False, server_default="5"),
        sa.Column("total_queries_used", sa.BigInteger, nullable=False, server_default="0"),
        sa.Column("total_tokens_used", sa.BigInteger, nullable=False, server_default="0"),
        sa.Column("total_conversations", sa.Integer, nullable=False, server_default="0"),
        sa.Column("last_query_date", sa.Date, nullable=True),
        sa.Column("last_query_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("last_conversation_started", sa.DateTime(timezone=True), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            onupdate=sa.func.now(),
        ),
    )

    op.create_unique_constraint("uq_ai_user_quotas_user_id", "ai_user_quotas", ["user_id"])
    op.create_index("ix_ai_user_quotas_user_id", "ai_user_quotas", ["user_id"])

    op.create_table(
        "ai_conversation_sessions",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("title", sa.String(500), nullable=False, server_default="New conversation"),
        sa.Column("language", sa.String(5), nullable=False, server_default="en"),
        sa.Column("tier_at_start", sa.String(20), nullable=False),
        sa.Column("current_tier", sa.String(20), nullable=False),
        sa.Column("status", sa.String(20), nullable=False, server_default="active"),
        sa.Column("context_summary", sa.Text, nullable=True),
        sa.Column("last_message_preview", sa.String(200), nullable=False, server_default=""),
        sa.Column("message_count", sa.Integer, nullable=False, server_default="0"),
        sa.Column("total_tokens_used", sa.BigInteger, nullable=False, server_default="0"),
        sa.Column("last_message_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            onupdate=sa.func.now(),
        ),
        sa.Column("deleted_at", sa.DateTime(timezone=True), nullable=True),
    )

    op.create_index("ix_ai_conversation_sessions_user_id", "ai_conversation_sessions", ["user_id"])
    op.create_index(
        "ix_ai_conversation_sessions_user_created",
        "ai_conversation_sessions",
        ["user_id", "created_at"],
    )
    op.create_index(
        "ix_ai_conversation_sessions_deleted_at", "ai_conversation_sessions", ["deleted_at"]
    )

    op.create_table(
        "ai_chat_messages",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column(
            "conversation_id",
            postgresql.UUID(as_uuid=True),
            sa.ForeignKey("ai_conversation_sessions.id", ondelete="CASCADE"),
            nullable=False,
        ),
        sa.Column("message_type", sa.String(20), nullable=False),
        sa.Column("user_query", sa.Text, nullable=True),
        sa.Column("query_language", sa.String(5), nullable=False, server_default="en"),
        sa.Column("query_embedding", Vector(1024), nullable=True),  # type: ignore[reportUnknownArgumentType]
        sa.Column(
            "retrieved_chunk_ids", postgresql.ARRAY(postgresql.UUID(as_uuid=True)), nullable=True
        ),
        sa.Column("context_chunks", postgresql.JSONB, nullable=True),
        sa.Column("llm_response", sa.Text, nullable=True),
        sa.Column("response_sources", postgresql.JSONB, nullable=True),
        sa.Column("processing_time_ms", sa.Integer, nullable=False, server_default="0"),
        sa.Column("token_usage", postgresql.JSONB, nullable=True),
        sa.Column("model_used", sa.String(100), nullable=False, server_default=""),
        sa.Column("prompt_version", sa.Integer, nullable=False, server_default="1"),
        sa.Column("trace_id", sa.String(255), nullable=True),
        sa.Column("cache_hit", sa.Boolean, nullable=False, server_default="false"),
        sa.Column("user_feedback", sa.Integer, nullable=True),
        sa.Column("feedback_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("is_context_cleared", sa.Boolean, nullable=False, server_default="false"),
        sa.Column("message_order", sa.Integer, nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            onupdate=sa.func.now(),
        ),
    )

    op.create_index("ix_ai_chat_messages_user_id", "ai_chat_messages", ["user_id"])
    op.create_index("ix_ai_chat_messages_conversation_id", "ai_chat_messages", ["conversation_id"])
    op.create_index(
        "ix_ai_chat_messages_conversation_order",
        "ai_chat_messages",
        ["conversation_id", "message_order"],
    )
    op.create_index("ix_ai_chat_messages_created_at", "ai_chat_messages", ["created_at"])


def downgrade() -> None:
    op.drop_table("ai_chat_messages")
    op.drop_table("ai_conversation_sessions")
    op.drop_table("ai_user_quotas")
    op.drop_table("document_chunks")
    op.drop_table("knowledge_documents")
    op.execute("DROP EXTENSION IF EXISTS vector")
