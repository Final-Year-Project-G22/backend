from __future__ import annotations

import asyncio
import uuid

import pytest

from core.domain.enums import Language, MessageType
from core.domain.models import AIChatMessage
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
