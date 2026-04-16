"""add embedding profiles

Revision ID: 003
Revises: 002
Create Date: 2026-04-16
"""

from collections.abc import Sequence

import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "003"
down_revision: str | None = "002"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "embedding_profiles",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("name", sa.String(100), nullable=False),
        sa.Column("model_name", sa.String(255), nullable=False),
        sa.Column("dimensions", sa.Integer, nullable=False),
        sa.Column("description", sa.String(500), nullable=True),
        sa.Column("tags", postgresql.ARRAY(sa.String(100)), nullable=True),
        sa.Column("is_active", sa.Boolean, nullable=False, server_default="true"),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            onupdate=sa.func.now(),
        ),
    )

    op.create_unique_constraint("uq_embedding_profiles_name", "embedding_profiles", ["name"])
    op.create_index("ix_embedding_profiles_name", "embedding_profiles", ["name"])
    op.create_index("ix_embedding_profiles_active", "embedding_profiles", ["is_active"])
    op.create_index(
        "ix_embedding_profiles_active_name", "embedding_profiles", ["is_active", "name"]
    )

    op.add_column(
        "document_chunks",
        sa.Column(
            "embedding_profile_id",
            postgresql.UUID(as_uuid=True),
            sa.ForeignKey("embedding_profiles.id", ondelete="SET NULL"),
            nullable=True,
        ),
    )
    op.create_index(
        "ix_document_chunks_embedding_profile_id",
        "document_chunks",
        ["embedding_profile_id"],
    )


def downgrade() -> None:
    op.drop_index("ix_document_chunks_embedding_profile_id", "document_chunks")
    op.drop_column("document_chunks", "embedding_profile_id")
    op.drop_table("embedding_profiles")
