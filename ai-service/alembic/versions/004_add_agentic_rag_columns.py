"""add tool_calls and agent_strategy columns to ai_chat_messages

Revision ID: 004
Revises: 003
Create Date: 2026-05-26
"""

from collections.abc import Sequence

import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "004"
down_revision: str | None = "003"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.add_column(
        "ai_chat_messages",
        sa.Column("tool_calls", postgresql.JSONB, nullable=True),
    )
    op.add_column(
        "ai_chat_messages",
        sa.Column(
            "agent_strategy",
            sa.String(20),
            nullable=False,
            server_default="simple",
        ),
    )


def downgrade() -> None:
    op.drop_column("ai_chat_messages", "agent_strategy")
    op.drop_column("ai_chat_messages", "tool_calls")
