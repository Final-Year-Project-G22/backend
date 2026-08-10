"""change embedding columns from 768d to 1024d for cohere embed-multilingual-v3.0

Cohere embed-multilingual-v3.0 always returns 1024-dimensional embeddings
(the API ignores output_dimension on this model), so the columns that were
narrowed to 768d for gemini text-multilingual-embedding-002 are widened back.

Revision ID: 006
Revises: 005
Create Date: 2026-08-11
"""

from collections.abc import Sequence

import sqlalchemy as sa
from pgvector.sqlalchemy import Vector

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "006"
down_revision: str | None = "005"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
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
        type_=Vector(1024),
        existing_type=Vector(768),
        existing_nullable=True,
    )


def downgrade() -> None:
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
