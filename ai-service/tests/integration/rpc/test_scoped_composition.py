from __future__ import annotations

import asyncio
import uuid
from unittest.mock import AsyncMock

import pytest

from core.domain.enums import Language, MessageType, Tier
from core.domain.models import AIChatMessage
from core.ports.core_service import CoreServicePort
from core.usecases.contracts import CreateSessionCommand
from infrastructure.database.connection import Database
from infrastructure.database.scoped import ScopedConversation, ScopedRepository


@pytest.mark.integration
async def test_conversation_creation_commits_quota_and_conversation(database: Database) -> None:
    user_id = uuid.uuid4()
    usecase = ScopedConversation(database, core_service=None)
    created = await usecase.create_session(
        CreateSessionCommand(user_id=user_id, title="New", language=Language.ENGLISH)
    )
    async with database.command() as reader:
        assert await reader.quota.get_quota(user_id) is not None
        assert await reader.conversation.get_session(created.id) is not None


@pytest.mark.integration
async def test_concurrent_repository_calls_do_not_share_a_session(database: Database) -> None:
    repository = ScopedRepository(database, "quota")
    first, second = uuid.uuid4(), uuid.uuid4()
    await asyncio.wait_for(
        asyncio.gather(repository.get_quota(first), repository.get_quota(second)), 5
    )
    async with database.command() as reader:
        assert await reader.quota.get_quota(first) is None
        assert await reader.quota.get_quota(second) is None


@pytest.mark.integration
async def test_slow_tier_lookup_does_not_hold_a_database_connection(database: Database) -> None:
    started, resume = asyncio.Event(), asyncio.Event()
    core = AsyncMock(spec=CoreServicePort)

    async def slow_tier(_user_id: uuid.UUID) -> Tier:
        started.set()
        await resume.wait()
        return Tier.PRO

    core.get_user_tier.side_effect = slow_tier
    user_id = uuid.uuid4()
    usecase = ScopedConversation(database, core_service=core)
    task = asyncio.create_task(
        usecase.create_session(
            CreateSessionCommand(user_id=user_id, title="Slow tier", language=Language.ENGLISH)
        )
    )
    try:
        await asyncio.wait_for(started.wait(), 5)
        assert database.engine.pool.checkedout() == 0
        resume.set()
        created = await asyncio.wait_for(task, 5)
        assert created.tier_at_start is Tier.PRO
        async with database.command() as reader:
            quota = await reader.quota.get_quota(user_id)
            assert quota is not None
            assert quota.tier is Tier.PRO
    finally:
        resume.set()
        if not task.done():
            task.cancel()
            await asyncio.gather(task, return_exceptions=True)


@pytest.mark.integration
async def test_two_concurrent_turns_append_in_order(database: Database) -> None:
    user_id = uuid.uuid4()
    usecase = ScopedConversation(database, core_service=None)
    conversation = await usecase.create_session(
        CreateSessionCommand(user_id=user_id, title="Shared", language=Language.ENGLISH)
    )

    def message(text: str) -> AIChatMessage:
        return AIChatMessage(
            user_id=user_id,
            conversation_id=conversation.id,
            message_type=MessageType.USER_QUERY,
            user_query=text,
            message_order=1,
        )

    await asyncio.gather(
        usecase.append_message(message("one")), usecase.append_message(message("two"))
    )
    messages = await usecase.list_messages(conversation.id)
    assert [item.message_order for item in messages] == [1, 2]
    assert {item.user_query for item in messages} == {"one", "two"}
