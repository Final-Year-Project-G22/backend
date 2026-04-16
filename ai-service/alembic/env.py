from __future__ import annotations

import asyncio
from logging.config import fileConfig

import pgvector  # noqa: F401 — required for Vector type deserialization
from sqlalchemy.engine import Connection

from alembic import context
from infrastructure.database.connection import Base, engine
from infrastructure.database.models_sqlalchemy import (  # noqa: F401 — registers models
    AIChatMessage,
    AIConversationSession,
    AIUserQuota,
    DocumentChunk,
    IngestionEventLedger,
    KnowledgeDocument,
)

config = context.config

if config.config_file_name is not None:
    fileConfig(config.config_file_name)

target_metadata = Base.metadata


def run_migrations_offline() -> None:
    from app.config import get_settings

    url = get_settings().DATABASE_URL
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )

    with context.begin_transaction():
        context.run_migrations()


def do_run_migrations(connection: Connection) -> None:
    context.configure(connection=connection, target_metadata=target_metadata)

    with context.begin_transaction():
        context.run_migrations()


async def run_async_migrations() -> None:
    connectable = engine

    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)

    await connectable.dispose()


def run_migrations_online() -> None:
    asyncio.run(run_async_migrations())


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
