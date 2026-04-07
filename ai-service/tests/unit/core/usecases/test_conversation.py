from __future__ import annotations

import uuid
from datetime import UTC, datetime
from unittest.mock import AsyncMock

import pytest

from core.domain.enums import Language, MessageType, SessionStatus, Tier
from core.domain.models import AIChatMessage, AIConversationSession, AIUserQuota
from core.usecases.contracts import CreateSessionCommand, ListSessionsQuery
from core.usecases.conversation import ConversationUseCase


def _build_session(*, user_id: uuid.UUID) -> AIConversationSession:
    now = datetime(2026, 4, 8, 9, 0, tzinfo=UTC)
    return AIConversationSession(
        user_id=user_id,
        title="Business support",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
        created_at=now,
        updated_at=now,
    )


@pytest.mark.asyncio
async def test_create_session_uses_quota_tier_when_available() -> None:
    user_id = uuid.uuid4()
    created_session = _build_session(user_id=user_id).model_copy(
        update={"tier_at_start": Tier.PRO, "current_tier": Tier.PRO}
    )

    conversation_repository = AsyncMock()
    conversation_repository.create_session.return_value = created_session

    quota_guard = AsyncMock()
    quota_guard.get_quota.return_value = AIUserQuota(user_id=user_id, tier=Tier.PRO)

    usecase = ConversationUseCase(conversation_repository, quota_guard=quota_guard)
    command = CreateSessionCommand(
        user_id=user_id,
        title="Help me register",
        language=Language.AMHARIC,
    )

    result = await usecase.create_session(command)

    assert result == created_session
    quota_guard.get_quota.assert_awaited_once_with(user_id)


@pytest.mark.asyncio
async def test_create_session_defaults_to_basic_tier_without_quota_guard() -> None:
    user_id = uuid.uuid4()
    created_session = _build_session(user_id=user_id)

    conversation_repository = AsyncMock()
    conversation_repository.create_session.return_value = created_session

    usecase = ConversationUseCase(conversation_repository)
    command = CreateSessionCommand(user_id=user_id, title="New session")

    result = await usecase.create_session(command)

    assert result.tier_at_start is Tier.BASIC
    assert result.current_tier is Tier.BASIC
    assert result.title == "Business support"
    assert result.user_id == user_id


@pytest.mark.asyncio
async def test_get_session_hides_deleted_by_default() -> None:
    user_id = uuid.uuid4()
    deleted_session = _build_session(user_id=user_id).model_copy(
        update={
            "status": SessionStatus.ARCHIVED,
            "deleted_at": datetime(2026, 4, 8, 10, 0, tzinfo=UTC),
        }
    )
    conversation_repository = AsyncMock()
    conversation_repository.get_session.return_value = deleted_session

    usecase = ConversationUseCase(conversation_repository)
    result = await usecase.get_session(deleted_session.id)

    assert result is None


@pytest.mark.asyncio
async def test_get_session_can_include_deleted() -> None:
    user_id = uuid.uuid4()
    deleted_session = _build_session(user_id=user_id).model_copy(
        update={
            "status": SessionStatus.ARCHIVED,
            "deleted_at": datetime(2026, 4, 8, 10, 0, tzinfo=UTC),
        }
    )
    conversation_repository = AsyncMock()
    conversation_repository.get_session.return_value = deleted_session

    usecase = ConversationUseCase(conversation_repository)
    result = await usecase.get_session(deleted_session.id, include_deleted=True)

    assert result == deleted_session


@pytest.mark.asyncio
async def test_list_sessions_delegates_with_query_values() -> None:
    user_id = uuid.uuid4()
    sessions = [_build_session(user_id=user_id)]

    conversation_repository = AsyncMock()
    conversation_repository.list_sessions_by_user.return_value = sessions

    usecase = ConversationUseCase(conversation_repository)
    query = ListSessionsQuery(user_id=user_id, limit=15, offset=5, include_deleted=True)
    result = await usecase.list_sessions(query)

    assert result == sessions
    conversation_repository.list_sessions_by_user.assert_awaited_once_with(
        user_id,
        limit=15,
        offset=5,
        include_deleted=True,
    )


@pytest.mark.asyncio
async def test_archive_session_uses_explicit_deleted_at() -> None:
    session_id = uuid.uuid4()
    deleted_at = datetime(2026, 4, 8, 12, 0, tzinfo=UTC)

    conversation_repository = AsyncMock()
    conversation_repository.soft_delete_session.return_value = True

    usecase = ConversationUseCase(conversation_repository)
    result = await usecase.archive_session(session_id, deleted_at=deleted_at)

    assert result is True
    conversation_repository.soft_delete_session.assert_awaited_once_with(
        session_id,
        deleted_at=deleted_at,
    )


@pytest.mark.asyncio
async def test_list_and_get_messages_delegate_to_repository() -> None:
    conversation_id = uuid.uuid4()
    message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=conversation_id,
        message_type=MessageType.AI_RESPONSE,
        llm_response="Use your license form",
        message_order=1,
    )

    conversation_repository = AsyncMock()
    conversation_repository.list_messages.return_value = [message]
    conversation_repository.get_message.return_value = message

    usecase = ConversationUseCase(conversation_repository)
    listed = await usecase.list_messages(conversation_id, limit=50, offset=10)
    found = await usecase.get_message(message.id)

    assert listed == [message]
    assert found == message
    conversation_repository.list_messages.assert_awaited_once_with(
        conversation_id,
        limit=50,
        offset=10,
    )
    conversation_repository.get_message.assert_awaited_once_with(message.id)
