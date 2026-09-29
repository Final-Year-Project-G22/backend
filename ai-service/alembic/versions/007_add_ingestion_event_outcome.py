"""Record the outcome of each ingestion event for safe completion recovery.

Revision ID: 007
Revises: 006
"""

from collections.abc import Sequence

import sqlalchemy as sa

from alembic import op

revision: str = "007"
down_revision: str | None = "006"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    # Legacy ledger rows do not prove completion. Keep their outcomes unknown.
    op.add_column("ingestion_event_ledger", sa.Column("outcome", sa.String(20), nullable=True))
    op.create_check_constraint(
        "ck_ingestion_event_ledger_outcome",
        "ingestion_event_ledger",
        "outcome IN ('completed', 'failed')",
    )


def downgrade() -> None:
    op.drop_constraint("ck_ingestion_event_ledger_outcome", "ingestion_event_ledger", type_="check")
    op.drop_column("ingestion_event_ledger", "outcome")
