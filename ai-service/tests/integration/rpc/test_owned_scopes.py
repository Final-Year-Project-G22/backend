from __future__ import annotations

import asyncio
import os
import uuid
from collections.abc import AsyncIterator

import asyncpg
import pytest
from sqlalchemy.engine import make_url

from core.domain.enums import Language, Tier
from core.domain.models import AIConversationSession, AIUserQuota
from infrastructure.database import models_sqlalchemy as models
from infrastructure.database.connection import Base, Database


@pytest.fixture
async def database() -> AsyncIterator[Database]:
    raw_url = os.environ.get("AI_TEST_DATABASE_URL")
    if raw_url is None:
        pytest.skip("AI_TEST_DATABASE_URL requires PostgreSQL with pgvector")
    url = make_url(raw_url)
    schema = f"ai247_{uuid.uuid4().hex}"
    admin = await asyncpg.connect(
        url.set(drivername="postgresql").render_as_string(hide_password=False)
    )
    try:
        await admin.execute("CREATE EXTENSION IF NOT EXISTS vector")
        await admin.execute(f'CREATE SCHEMA "{schema}"')
        db = Database(
            raw_url, connect_args={"server_settings": {"search_path": f"{schema},public"}}
        )
        try:
            async with db.engine.begin() as connection:
                await connection.run_sync(
                    Base.metadata.create_all,
                    tables=[
                        models.AIUserQuota.__table__,
                        models.AIConversationSession.__table__,
                        models.AIChatMessage.__table__,
                        models.IngestionEventLedger.__table__,
                    ],
                )
            yield db
        finally:
            await db.engine.dispose()
            await admin.execute(f'DROP SCHEMA "{schema}" CASCADE')
    finally:
        await admin.close()


@pytest.mark.integration
async def test_command_commits_related_writes_and_new_connection_sees_them(
    database: Database,
) -> None:
    user_id = uuid.uuid4()
    conversation = AIConversationSession(
        user_id=user_id,
        title="First turn",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
    )
    async with database.command() as command:
        await command.quota.upsert_quota(AIUserQuota(user_id=user_id))
        await command.conversation.create_session(conversation)

    async with database.command() as reader:
        assert await reader.quota.get_quota(user_id) is not None
        assert await reader.conversation.get_session(conversation.id) is not None


@pytest.mark.integration
async def test_failed_command_rolls_back_related_writes(database: Database) -> None:
    user_id = uuid.uuid4()
    conversation = AIConversationSession(
        user_id=user_id,
        title="Failed turn",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
    )

    async def fail() -> None:
        async with database.command() as command:
            await command.quota.upsert_quota(AIUserQuota(user_id=user_id))
            await command.conversation.create_session(conversation)
            raise RuntimeError("stop")

    with pytest.raises(RuntimeError, match="stop"):
        await fail()

    async with database.command() as reader:
        assert await reader.quota.get_quota(user_id) is None
        assert await reader.conversation.get_session(conversation.id) is None


@pytest.mark.integration
async def test_cancelled_command_releases_connection(database: Database) -> None:
    started = asyncio.Event()

    async def cancelled() -> None:
        async with database.command() as command:
            await command.quota.upsert_quota(AIUserQuota(user_id=uuid.uuid4()))
            started.set()
            await asyncio.Event().wait()

    task = asyncio.create_task(cancelled())
    await asyncio.wait_for(started.wait(), 5)
    task.cancel()
    with pytest.raises(asyncio.CancelledError):
        await task
    async with asyncio.timeout(5):
        async with database.command() as command:
            await command.quota.upsert_quota(AIUserQuota(user_id=uuid.uuid4()))
