from __future__ import annotations

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.orm import DeclarativeBase


class Base(DeclarativeBase):
    pass


class DatabaseCommand:
    """Repositories in a command use exactly one session and transaction."""

    def __init__(self, session: AsyncSession) -> None:
        from infrastructure.database.repositories import (  # noqa: PLC0415 - Base must exist before models import
            SqlAlchemyConversationRepository,
            SqlAlchemyIngestionEventLedgerRepository,
            SqlAlchemyKnowledgeRepository,
            SqlAlchemyQuotaRepository,
        )

        self.conversation = SqlAlchemyConversationRepository(session)
        self.quota = SqlAlchemyQuotaRepository(session)
        self.knowledge = SqlAlchemyKnowledgeRepository(session)
        self.ledger = SqlAlchemyIngestionEventLedgerRepository(session)


class Database:
    def __init__(
        self, url: str, *, connect_args: dict[str, object] | None = None, debug: bool = False
    ) -> None:
        self.engine: AsyncEngine = create_async_engine(
            url,
            pool_size=10,
            max_overflow=20,
            pool_pre_ping=True,
            echo=debug,
            connect_args=connect_args or {},
        )
        self.sessions = async_sessionmaker(self.engine, expire_on_commit=False)

    @asynccontextmanager
    async def command(self) -> AsyncIterator[DatabaseCommand]:
        async with self.sessions() as session, session.begin():
            yield DatabaseCommand(session)
