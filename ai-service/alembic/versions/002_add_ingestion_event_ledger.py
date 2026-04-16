"""add ingestion event ledger

Revision ID: 002
Revises: 001
Create Date: 2026-04-15
"""

from collections.abc import Sequence

import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "002"
down_revision: str | None = "001"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "ingestion_event_ledger",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("event_id", sa.String(100), nullable=False),
        sa.Column("idempotency_key", sa.String(255), nullable=False),
        sa.Column("account_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("document_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("occurred_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
    )

    op.create_unique_constraint(
        "uq_ingestion_event_ledger_event_id",
        "ingestion_event_ledger",
        ["event_id"],
    )
    op.create_unique_constraint(
        "uq_ingestion_event_ledger_idempotency_key",
        "ingestion_event_ledger",
        ["idempotency_key"],
    )
    op.create_index("ix_ingestion_event_ledger_event_id", "ingestion_event_ledger", ["event_id"])
    op.create_index(
        "ix_ingestion_event_ledger_idempotency_key",
        "ingestion_event_ledger",
        ["idempotency_key"],
    )
    op.create_index(
        "ix_ingestion_event_ledger_account_id",
        "ingestion_event_ledger",
        ["account_id"],
    )
    op.create_index(
        "ix_ingestion_event_ledger_document_id",
        "ingestion_event_ledger",
        ["document_id"],
    )
    op.create_index(
        "ix_ingestion_event_ledger_document_occurred",
        "ingestion_event_ledger",
        ["document_id", "occurred_at"],
    )


def downgrade() -> None:
    op.drop_table("ingestion_event_ledger")
