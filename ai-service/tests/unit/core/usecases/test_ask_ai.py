from __future__ import annotations

import uuid
from datetime import UTC, datetime
from unittest.mock import AsyncMock

import pytest

from core.domain.enums import Language, MessageType, Tier
from core.domain.models import AIChatMessage, AIConversationSession
from core.usecases.ask_ai import AskAIUseCase
from core.usecases.contracts import AskAICommand, AskAIResult


def _build_conversation() -> AIConversationSession:
    now = datetime(2026, 4, 8, 10, 0, tzinfo=UTC)
    return AIConversationSession(
        user_id=uuid.uuid4(),
        title="Test",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
        created_at=now,
        updated_at=now,
    )


def _build_ai_message(conversation_id: uuid.UUID) -> AIChatMessage:
    now = datetime(2026, 4, 8, 10, 2, tzinfo=UTC)
    return AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=conversation_id,
        message_type=MessageType.AI_RESPONSE,
        llm_response="Answer",
        message_order=2,
        created_at=now,
        updated_at=now,
    )


def _build_command(**overrides: object) -> AskAICommand:
    payload: dict[str, object] = {
        "user_id": uuid.uuid4(),
        "account_id": uuid.uuid4(),
        "prompt": "How do I register?",
        "language": Language.ENGLISH,
        "conversation_id": uuid.uuid4(),
    }
    payload.update(overrides)
    return AskAICommand(**payload)  # type: ignore[arg-type]


def _make_strategy() -> AsyncMock:
    strategy = AsyncMock()
    strategy.execute = AsyncMock()
    strategy.execute_stream = AsyncMock()
    return strategy


@pytest.mark.asyncio
async def test_execute_delegates_to_simple_strategy() -> None:
    strategy = _make_strategy()
    usecase = AskAIUseCase(simple_strategy=strategy)
    command = _build_command()
    await usecase.execute(command)
    strategy.execute.assert_awaited_once_with(command)


@pytest.mark.asyncio
async def test_execute_delegates_to_agentic_strategy_when_requested() -> None:
    simple = _make_strategy()
    agentic = _make_strategy()
    usecase = AskAIUseCase(simple_strategy=simple, agentic_strategy=agentic)
    command = _build_command(strategy="agentic")
    await usecase.execute(command)
    agentic.execute.assert_awaited_once_with(command)
    simple.execute.assert_not_awaited()


@pytest.mark.asyncio
async def test_execute_falls_back_to_simple_when_no_agentic() -> None:
    simple = _make_strategy()
    usecase = AskAIUseCase(simple_strategy=simple)
    command = _build_command(strategy="agentic")
    await usecase.execute(command)
    simple.execute.assert_awaited_once_with(command)


@pytest.mark.asyncio
async def test_execute_returns_strategy_result() -> None:
    conversation = _build_conversation()
    ai_message = _build_ai_message(conversation.id)
    result = AskAIResult(
        conversation=conversation,
        user_message=AIChatMessage(
            user_id=uuid.uuid4(),
            conversation_id=conversation.id,
            message_type=MessageType.USER_QUERY,
            user_query="test",
            query_language=Language.ENGLISH,
            message_order=1,
        ),
        ai_message=ai_message,
    )
    strategy = _make_strategy()
    strategy.execute.return_value = result
    usecase = AskAIUseCase(simple_strategy=strategy)
    command = _build_command()
    response = await usecase.execute(command)
    assert response is result
