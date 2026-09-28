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

from app.config import get_settings


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


settings = get_settings()
engine = create_async_engine(
    settings.DATABASE_URL,
    pool_size=10,
    max_overflow=20,
    pool_pre_ping=True,
    echo=settings.DEBUG,
)

async_session_factory = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)


async def get_session() -> AsyncIterator[AsyncSession]:
    async with async_session_factory() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise


@asynccontextmanager
async def session_scope() -> AsyncIterator[AsyncSession]:
    """Provide a transactional scope around a series of operations."""
    async with async_session_factory() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
