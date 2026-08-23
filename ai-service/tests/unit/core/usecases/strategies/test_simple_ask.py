from __future__ import annotations

import uuid
from collections.abc import AsyncIterator
from unittest.mock import AsyncMock, MagicMock

import pytest

from core.domain.enums import DocumentSource, Language, MessageType, Tier
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.stream_events import AskStreamEventType
from core.domain.value_objects import SearchHit
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.embedding import EmbeddingPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from core.ports.llm import LLMChunk, LLMPort
from core.usecases.contracts import AskAICommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.strategies.simple_ask import SimpleAskStrategy
from infrastructure.prompts import PromptLoader
from tests.unit.core.usecases.strategies._message_store import _MessageStore


def _make_strategy(store: _MessageStore) -> tuple[SimpleAskStrategy, MagicMock, AskAICommand]:
    user_id = uuid.uuid4()
    command = AskAICommand(
        user_id=user_id,
        prompt="What is required for a trade licence?",
        language=Language.AMHARIC,
        strategy="simple",
    )

    session = AIConversationSession(
        user_id=user_id,
        title=command.prompt[:80],
        language=command.language,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
    )
    repository = MagicMock(spec=ConversationRepositoryPort)
    repository.create_session = AsyncMock(return_value=session)
    repository.get_session = AsyncMock(return_value=None)
    repository.list_messages = AsyncMock(side_effect=store.list)
    repository.add_message = AsyncMock(side_effect=store.add)
    repository.update_message = AsyncMock(side_effect=store.update)
    conversation = ConversationUseCase(repository)

    hit = SearchHit(
        chunk_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        score=0.95,
        chunk_text="Amharic trade licence context",
        chunk_index=0,
        source=DocumentSource.GOVERNMENT,
        language=Language.AMHARIC,
        document_title="Trade Licence Guide",
    )
    knowledge_repository = MagicMock(spec=KnowledgeRepositoryPort)
    knowledge_repository.search_vector = AsyncMock(return_value=[hit])
    knowledge_repository.search_bm25 = AsyncMock(return_value=[])

    embedding_port = MagicMock(spec=EmbeddingPort)
    embedding_port.embed_query = AsyncMock(return_value=[1.0, 0.0])

    llm_port = MagicMock(spec=LLMPort)
    llm_port.generate_stream = _fake_stream

    prompt_loader = MagicMock(spec=PromptLoader)
    prompt_loader.render_simple.return_value = "simple system prompt"

    strategy = SimpleAskStrategy(
        conversation=conversation,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
        prompt_loader=prompt_loader,
    )
    return strategy, repository, command


async def _fake_stream(*_args: object, **_kwargs: object) -> AsyncIterator[LLMChunk]:
    yield LLMChunk(text="Grounded answer")


def _last_message(*, message_order: int) -> AIChatMessage:
    return AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Previous answer",
        message_order=message_order,
    )


@pytest.mark.asyncio
async def test_persisted_messages_follow_most_recent_order_across_turns() -> None:
    store = _MessageStore(_last_message(message_order=5))
    strategy, _repository, command = _make_strategy(store)

    events = [event async for event in strategy.execute_stream(command)]
    assert events[-1].type is AskStreamEventType.DONE
    assert [m.message_order for m in store.messages] == [5, 6, 7]
    assert store.messages[1].message_type is MessageType.USER_QUERY
    assert store.messages[2].message_type is MessageType.AI_RESPONSE

    events = [event async for event in strategy.execute_stream(command)]
    assert events[-1].type is AskStreamEventType.DONE
    assert [m.message_order for m in store.messages] == [5, 6, 7, 8, 9]


@pytest.mark.asyncio
async def test_first_message_in_empty_conversation_gets_order_one() -> None:
    store = _MessageStore()
    strategy, _repository, command = _make_strategy(store)

    events = [event async for event in strategy.execute_stream(command)]
    assert events[-1].type is AskStreamEventType.DONE
    assert [m.message_order for m in store.messages] == [1, 2]
