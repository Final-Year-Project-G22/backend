from __future__ import annotations

import uuid
from datetime import UTC, datetime
from unittest.mock import AsyncMock, Mock

import pytest
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.enums import Language, MessageType, SessionStatus, Tier
from core.domain.exceptions import RepositoryError
from core.domain.models import AIChatMessage, AIConversationSession
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.conversation_repository import (
    SqlAlchemyConversationRepository,
)
from infrastructure.database.repositories.mappers import to_orm_message, to_orm_session


class _ScalarResult:
    def __init__(
        self, value: sa_models.AIConversationSession | sa_models.AIChatMessage | None
    ) -> None:
        self._value = value

    def scalar_one_or_none(
        self,
    ) -> sa_models.AIConversationSession | sa_models.AIChatMessage | None:
        return self._value


class _ScalarsResult:
    def __init__(
        self, values: list[sa_models.AIConversationSession | sa_models.AIChatMessage]
    ) -> None:
        self._values = values

    def all(self) -> list[sa_models.AIConversationSession | sa_models.AIChatMessage]:
        return self._values


class _ListResult:
    def __init__(
        self, values: list[sa_models.AIConversationSession | sa_models.AIChatMessage]
    ) -> None:
        self._values = values

    def scalars(self) -> _ScalarsResult:
        return _ScalarsResult(self._values)


def _build_session_mock() -> AsyncMock:
    session = AsyncMock(spec=AsyncSession)
    session.add = Mock()
    return session


def _build_domain_session() -> AIConversationSession:
    now = datetime(2026, 4, 7, 10, 0, tzinfo=UTC)
    return AIConversationSession(
        user_id=uuid.uuid4(),
        title="Business setup",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
        status=SessionStatus.ACTIVE,
        message_count=1,
        last_message_at=now,
        created_at=now,
        updated_at=now,
    )


def _build_domain_message(*, conversation_id: uuid.UUID) -> AIChatMessage:
    now = datetime(2026, 4, 7, 10, 5, tzinfo=UTC)
    return AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=conversation_id,
        message_type=MessageType.AI_RESPONSE,
        llm_response="Start with registration.",
        model_used="gemini-1.5-flash",
        message_order=1,
        created_at=now,
        updated_at=now,
    )


@pytest.mark.asyncio
async def test_create_session_adds_and_returns_domain_model() -> None:
    session = _build_session_mock()
    repo = SqlAlchemyConversationRepository(session)
    domain_session = _build_domain_session()

    created = await repo.create_session(domain_session)

    assert created == domain_session
    session.add.assert_called_once()
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_get_session_returns_none_when_missing() -> None:
    session = _build_session_mock()
    session.execute.return_value = _ScalarResult(None)
    repo = SqlAlchemyConversationRepository(session)

    found = await repo.get_session(uuid.uuid4())

    assert found is None


@pytest.mark.asyncio
async def test_update_session_raises_repository_error_if_missing() -> None:
    session = _build_session_mock()
    session.execute.return_value = _ScalarResult(None)
    repo = SqlAlchemyConversationRepository(session)

    with pytest.raises(RepositoryError, match="conversation session not found"):
        await repo.update_session(_build_domain_session())


@pytest.mark.asyncio
async def test_list_sessions_by_user_applies_deleted_filter_by_default() -> None:
    domain_session = _build_domain_session()
    orm_session = to_orm_session(domain_session)

    session = _build_session_mock()
    session.execute.return_value = _ListResult([orm_session])
    repo = SqlAlchemyConversationRepository(session)

    sessions = await repo.list_sessions_by_user(domain_session.user_id)

    assert sessions == [domain_session]
    statement = session.execute.await_args.args[0]
    assert "deleted_at IS NULL" in str(statement)


@pytest.mark.asyncio
async def test_soft_delete_session_marks_session_as_archived() -> None:
    domain_session = _build_domain_session()
    orm_session = to_orm_session(domain_session)

    session = _build_session_mock()
    session.execute.return_value = _ScalarResult(orm_session)
    repo = SqlAlchemyConversationRepository(session)

    deleted_at = datetime(2026, 4, 7, 11, 0, tzinfo=UTC)
    deleted = await repo.soft_delete_session(domain_session.id, deleted_at=deleted_at)

    assert deleted is True
    assert orm_session.status == SessionStatus.ARCHIVED.value
    assert orm_session.deleted_at == deleted_at
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_add_and_get_message_round_trip() -> None:
    domain_session = _build_domain_session()
    domain_message = _build_domain_message(conversation_id=domain_session.id)
    orm_message = to_orm_message(domain_message)

    session = _build_session_mock()
    session.execute.return_value = _ScalarResult(orm_message)
    repo = SqlAlchemyConversationRepository(session)

    created = await repo.add_message(domain_message)
    found = await repo.get_message(domain_message.id)

    assert created == domain_message
    assert found == domain_message
    session.add.assert_called_once()


@pytest.mark.asyncio
async def test_list_messages_returns_domain_models() -> None:
    domain_session = _build_domain_session()
    domain_message = _build_domain_message(conversation_id=domain_session.id)

    session = _build_session_mock()
    session.execute.return_value = _ListResult([to_orm_message(domain_message)])
    repo = SqlAlchemyConversationRepository(session)

    messages = await repo.list_messages(domain_session.id)

    assert messages == [domain_message]


@pytest.mark.asyncio
async def test_list_messages_descending_orders_most_recent_first() -> None:
    domain_session = _build_domain_session()
    domain_message = _build_domain_message(conversation_id=domain_session.id)

    session = _build_session_mock()
    session.execute.return_value = _ListResult([to_orm_message(domain_message)])
    repo = SqlAlchemyConversationRepository(session)

    messages = await repo.list_messages(domain_session.id, limit=1, descending=True)

    assert messages == [domain_message]
    statement = session.execute.await_args.args[0]
    compiled = str(statement)
    assert "message_order DESC" in compiled
    assert "message_order ASC" not in compiled


@pytest.mark.asyncio
async def test_get_session_wraps_sqlalchemy_errors() -> None:
    session = _build_session_mock()
    session.execute.side_effect = SQLAlchemyError("boom")
    repo = SqlAlchemyConversationRepository(session)

    with pytest.raises(RepositoryError, match="failed to fetch conversation session"):
        await repo.get_session(uuid.uuid4())
