"""change embedding columns from 1024d to 768d for text-multilingual-embedding-002

Revision ID: 005
Revises: 004
Create Date: 2026-05-26
"""

from collections.abc import Sequence

import sqlalchemy as sa
from pgvector.sqlalchemy import Vector  # type: ignore[import-untyped]  # pgvector ships no py.typed

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "005"
down_revision: str | None = "004"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.alter_column(
        "document_chunks",
        "embedding",
        type_=Vector(768),
        existing_type=Vector(1024),
        existing_nullable=True,
    )
    op.alter_column(
        "ai_chat_messages",
        "query_embedding",
        type_=Vector(768),
        existing_type=Vector(1024),
        existing_nullable=True,
    )


def downgrade() -> None:
    op.alter_column(
        "document_chunks",
        "embedding",
        type_=Vector(1024),
        existing_type=Vector(768),
        existing_nullable=True,
    )
    op.alter_column(
        "ai_chat_messages",
        "query_embedding",
        type_=Vector(768),
        existing_type=Vector(1024),
        existing_nullable=True,
    )
