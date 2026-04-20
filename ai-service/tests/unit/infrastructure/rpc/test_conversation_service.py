from __future__ import annotations

import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

import grpc
from core.domain.enums import Language, SessionStatus
from infrastructure.rpc.services.conversation_service import AIConversationService


def _make_list_request(**overrides: object) -> SimpleNamespace:
    payload: dict[str, object] = {
        "user_id": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
        "limit": 20,
        "offset": 0,
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


def _make_get_request(**overrides: object) -> SimpleNamespace:
    payload: dict[str, object] = {
        "session_id": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
        "message_limit": 50,
        "message_offset": 0,
        "include_deleted": False,
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


def _make_archive_request(**overrides: object) -> SimpleNamespace:
    payload: dict[str, object] = {
        "session_id": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


@pytest.mark.asyncio
async def test_list_conversations_returns_sessions() -> None:
    usecase = AsyncMock()
    session_id = uuid.uuid4()
    user_id = uuid.uuid4()
    usecase.list_sessions.return_value = [
        SimpleNamespace(
            id=session_id,
            user_id=user_id,
            title="Test conversation",
            language=Language.ENGLISH,
            created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T00:00:00"),
            updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T00:00:00"),
        )
    ]

    service = AIConversationService(conversation_usecase=usecase)
    request = _make_list_request(user_id=str(user_id), account_id=str(uuid.uuid4()))
    context = SimpleNamespace(abort=AsyncMock())

    response = await service.ListConversations(request, context)

    assert len(response.sessions) == 1
    assert response.sessions[0].title == "Test conversation"
    usecase.list_sessions.assert_called_once()


@pytest.mark.asyncio
async def test_list_conversations_invalid_user_id_aborts() -> None:
    usecase = AsyncMock()
    service = AIConversationService(conversation_usecase=usecase)

    request = _make_list_request(user_id="invalid-uuid")
    abort = AsyncMock()
    context = SimpleNamespace(abort=abort)

    await service.ListConversations(request, context)

    abort.assert_called_once_with(
        grpc.StatusCode.INVALID_ARGUMENT,
        "Invalid UUID or enum: badly formed hexadecimal UUID string",
    )


@pytest.mark.asyncio
async def test_get_conversation_returns_session_and_messages() -> None:
    usecase = AsyncMock()
    session_id = uuid.uuid4()
    account_id = uuid.uuid4()
    message_id = uuid.uuid4()

    usecase.get_session.return_value = SimpleNamespace(
        id=session_id,
        title="Test session",
        language=Language.ENGLISH,
        status=SessionStatus.ACTIVE,
        message_count=1,
        created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T00:00:00"),
        updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T00:00:00"),
    )
    usecase.list_messages.return_value = [
        SimpleNamespace(
            id=message_id,
            message_type=SimpleNamespace(value="user_query"),
            user_query="Hello",
            llm_response=None,
            retrieved_chunk_ids=[],
            token_usage=None,
            created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T00:00:00"),
        )
    ]

    service = AIConversationService(conversation_usecase=usecase)
    request = _make_get_request(session_id=str(session_id), account_id=str(account_id))
    context = SimpleNamespace(abort=AsyncMock())

    response = await service.GetConversation(request, context)

    assert response.session.title == "Test session"
    assert response.total_messages == 1


@pytest.mark.asyncio
async def test_get_conversation_not_found_aborts() -> None:
    usecase = AsyncMock()
    usecase.get_session.return_value = None

    service = AIConversationService(conversation_usecase=usecase)
    request = _make_get_request(session_id=str(uuid.uuid4()))
    abort = AsyncMock()
    context = SimpleNamespace(abort=abort)

    await service.GetConversation(request, context)

    abort.assert_called_once_with(grpc.StatusCode.NOT_FOUND, "Session not found")


@pytest.mark.asyncio
async def test_archive_conversation_returns_success() -> None:
    usecase = AsyncMock()
    session_id = uuid.uuid4()
    usecase.get_session.return_value = SimpleNamespace(id=session_id)
    usecase.archive_session.return_value = True

    service = AIConversationService(conversation_usecase=usecase)
    request = _make_archive_request(session_id=str(session_id))
    context = SimpleNamespace(abort=AsyncMock())

    response = await service.ArchiveConversation(request, context)

    assert response.success is True
    usecase.archive_session.assert_called_once_with(session_id)


@pytest.mark.asyncio
async def test_archive_conversation_not_found_aborts() -> None:
    usecase = AsyncMock()
    usecase.get_session.return_value = None

    service = AIConversationService(conversation_usecase=usecase)
    request = _make_archive_request(session_id=str(uuid.uuid4()))
    abort = AsyncMock()
    context = SimpleNamespace(abort=abort)

    await service.ArchiveConversation(request, context)

    abort.assert_called_once_with(grpc.StatusCode.NOT_FOUND, "Session not found")
