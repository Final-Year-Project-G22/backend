from __future__ import annotations

import uuid
from datetime import UTC, datetime
from typing import cast
from unittest.mock import AsyncMock

import pytest

from core.domain.enums import DocumentSource, Language, MessageType, Tier
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.value_objects import SearchHit
from core.usecases.ask_ai import AskAIUseCase
from core.usecases.contracts import AskAICommand, AskAIResult
from core.usecases.conversation import ConversationUseCase


def _build_conversation(*, user_id: uuid.UUID) -> AIConversationSession:
    now = datetime(2026, 4, 8, 10, 0, tzinfo=UTC)
    return AIConversationSession(
        user_id=user_id,
        title="Test conversation",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
        created_at=now,
        updated_at=now,
    )


def _build_search_hit(*, chunk_id: uuid.UUID | None = None) -> SearchHit:
    return SearchHit(
        chunk_id=chunk_id or uuid.uuid4(),
        document_id=uuid.uuid4(),
        score=0.85,
        chunk_text="Bring your ID to the office.",
        chunk_index=0,
        source=DocumentSource.GUIDE,
        language=Language.ENGLISH,
    )


def _build_user_message(*, conversation_id: uuid.UUID) -> AIChatMessage:
    now = datetime(2026, 4, 8, 10, 1, tzinfo=UTC)
    return AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=conversation_id,
        message_type=MessageType.USER_QUERY,
        user_query="How do I register?",
        query_language=Language.ENGLISH,
        message_order=1,
        created_at=now,
        updated_at=now,
    )


def _build_ai_message(
    *,
    conversation_id: uuid.UUID,
    llm_response: str = "Go to the office.",
    cache_hit: bool = False,
) -> AIChatMessage:
    now = datetime(2026, 4, 8, 10, 2, tzinfo=UTC)
    return AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=conversation_id,
        message_type=MessageType.AI_RESPONSE,
        llm_response=llm_response,
        query_language=Language.ENGLISH,
        message_order=2,
        cache_hit=cache_hit,
        created_at=now,
        updated_at=now,
    )


def _build_command(*, user_id: uuid.UUID, conversation_id: uuid.UUID | None = None) -> AskAICommand:
    return AskAICommand(
        user_id=user_id,
        prompt="How do I register my business?",
        language=Language.ENGLISH,
        conversation_id=conversation_id,
        vector_top_k=3,
        bm25_top_k=3,
    )


def _make_conversation_usecase(
    *,
    conversation: AIConversationSession,
    user_message: AIChatMessage,
    ai_message: AIChatMessage,
    messages_list: list[AIChatMessage] | None = None,
) -> tuple[ConversationUseCase, AsyncMock]:
    conversation_usecase = AsyncMock()
    conversation_usecase.get_session.return_value = conversation
    conversation_usecase.create_session.return_value = conversation
    conversation_usecase.list_messages.return_value = messages_list or []

    conversation_repository = AsyncMock()
    conversation_repository.add_message.side_effect = [
        user_message,
        user_message,
        ai_message,
    ]
    conversation_usecase.conversation_repository = conversation_repository

    return conversation_usecase, conversation_repository


@pytest.mark.asyncio
async def test_execute_creates_new_conversation_when_none_provided() -> None:
    user_id = uuid.uuid4()
    conversation = _build_conversation(user_id=user_id)
    user_message = _build_user_message(conversation_id=conversation.id)
    ai_message = _build_ai_message(conversation_id=conversation.id)
    search_hit = _build_search_hit()

    conversation_usecase, _ = _make_conversation_usecase(
        conversation=conversation,
        user_message=user_message,
        ai_message=ai_message,
    )

    quota_guard = AsyncMock()

    knowledge_repository = AsyncMock()
    knowledge_repository.search_vector.return_value = [search_hit]
    knowledge_repository.search_bm25.return_value = []

    embedding_port = AsyncMock()
    embedding_port.embed_query.return_value = [0.1, 0.2, 0.3]

    llm_port = AsyncMock()
    llm_port.generate.return_value = "Go to the office."

    usecase = AskAIUseCase(
        conversation=conversation_usecase,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
    )

    command = _build_command(user_id=user_id, conversation_id=None)
    result = await usecase.execute(command)

    assert isinstance(result, AskAIResult)
    assert result.conversation == conversation
    assert result.ai_message.llm_response == "Go to the office."
    assert result.retrieved_hits == [search_hit]
    assert result.cache_hit is False
    cast(AsyncMock, conversation_usecase.create_session).assert_awaited_once()


@pytest.mark.asyncio
async def test_execute_uses_existing_conversation_when_provided() -> None:
    user_id = uuid.uuid4()
    conversation_id = uuid.uuid4()
    conversation = _build_conversation(user_id=user_id)
    conversation.id = conversation_id
    user_message = _build_user_message(conversation_id=conversation_id)
    ai_message = _build_ai_message(conversation_id=conversation_id)

    conversation_usecase, _ = _make_conversation_usecase(
        conversation=conversation,
        user_message=user_message,
        ai_message=ai_message,
    )

    quota_guard = AsyncMock()

    knowledge_repository = AsyncMock()
    knowledge_repository.search_vector.return_value = []
    knowledge_repository.search_bm25.return_value = []

    embedding_port = AsyncMock()
    embedding_port.embed_query.return_value = [0.1, 0.2, 0.3]

    llm_port = AsyncMock()
    llm_port.generate.return_value = "Visit the registration office."

    usecase = AskAIUseCase(
        conversation=conversation_usecase,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
    )

    command = _build_command(user_id=user_id, conversation_id=conversation_id)
    result = await usecase.execute(command)

    assert result.conversation.id == conversation_id
    cast(AsyncMock, conversation_usecase.get_session).assert_awaited_once_with(conversation_id)
    cast(AsyncMock, conversation_usecase.create_session).assert_not_awaited()


@pytest.mark.asyncio
async def test_execute_uses_custom_title_for_new_conversation() -> None:
    user_id = uuid.uuid4()
    conversation = _build_conversation(user_id=user_id)
    user_message = _build_user_message(conversation_id=conversation.id)
    ai_message = _build_ai_message(conversation_id=conversation.id)

    conversation_usecase, _ = _make_conversation_usecase(
        conversation=conversation,
        user_message=user_message,
        ai_message=ai_message,
    )

    quota_guard = AsyncMock()
    knowledge_repository = AsyncMock()
    knowledge_repository.search_vector.return_value = []
    knowledge_repository.search_bm25.return_value = []
    embedding_port = AsyncMock()
    embedding_port.embed_query.return_value = [0.1, 0.2, 0.3]
    llm_port = AsyncMock()
    llm_port.generate.return_value = "Visit the office."

    usecase = AskAIUseCase(
        conversation=conversation_usecase,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
    )

    command = AskAICommand(
        user_id=user_id,
        prompt="How do I register my business?",
        title="My Custom Title",
        language=Language.ENGLISH,
        conversation_id=None,
        vector_top_k=3,
        bm25_top_k=3,
    )
    await usecase.execute(command)

    create_args = cast(AsyncMock, conversation_usecase.create_session).await_args.args
    create_cmd = create_args[0]
    assert create_cmd.title == "My Custom Title"


@pytest.mark.asyncio
async def test_execute_returns_cache_hit_when_cached() -> None:
    user_id = uuid.uuid4()
    conversation = _build_conversation(user_id=user_id)
    user_message = _build_user_message(conversation_id=conversation.id)
    ai_message = _build_ai_message(conversation_id=conversation.id, cache_hit=True)

    conversation_usecase = AsyncMock()
    conversation_usecase.get_session.return_value = conversation
    conversation_usecase.create_session.return_value = conversation
    conversation_usecase.list_messages.return_value = []

    conversation_repository = AsyncMock()
    conversation_repository.add_message.side_effect = [user_message, ai_message]
    conversation_usecase.conversation_repository = conversation_repository

    quota_guard = AsyncMock()

    cache = AsyncMock()
    cache.get.return_value = "Cached response"

    embedding_port = AsyncMock()
    llm_port = AsyncMock()
    knowledge_repository = AsyncMock()

    usecase = AskAIUseCase(
        conversation=conversation_usecase,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
        cache=cache,
    )

    command = _build_command(user_id=user_id, conversation_id=conversation.id)
    result = await usecase.execute(command)

    assert result.cache_hit is True
    assert result.ai_message.cache_hit is True
    embedding_port.embed_query.assert_not_awaited()
    llm_port.generate.assert_not_awaited()
    knowledge_repository.search_vector.assert_not_awaited()
    knowledge_repository.search_bm25.assert_not_awaited()


@pytest.mark.asyncio
async def test_execute_merges_and_deduplicates_hits() -> None:
    user_id = uuid.uuid4()
    conversation = _build_conversation(user_id=user_id)
    user_message = _build_user_message(conversation_id=conversation.id)
    ai_message = _build_ai_message(conversation_id=conversation.id)

    shared_chunk_id = uuid.uuid4()
    vector_hit = _build_search_hit(chunk_id=shared_chunk_id)
    bm25_hit = _build_search_hit(chunk_id=shared_chunk_id)
    unique_bm25_hit = _build_search_hit()
    expected_hits = 2

    conversation_usecase, _ = _make_conversation_usecase(
        conversation=conversation,
        user_message=user_message,
        ai_message=ai_message,
    )

    quota_guard = AsyncMock()

    knowledge_repository = AsyncMock()
    knowledge_repository.search_vector.return_value = [vector_hit]
    knowledge_repository.search_bm25.return_value = [bm25_hit, unique_bm25_hit]

    embedding_port = AsyncMock()
    embedding_port.embed_query.return_value = [0.1, 0.2, 0.3]

    llm_port = AsyncMock()
    llm_port.generate.return_value = "Answer"

    usecase = AskAIUseCase(
        conversation=conversation_usecase,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
    )

    command = _build_command(user_id=user_id, conversation_id=conversation.id)
    result = await usecase.execute(command)

    assert len(result.retrieved_hits) == expected_hits
    assert result.retrieved_hits[0].chunk_id == shared_chunk_id
    assert result.retrieved_hits[1].chunk_id == unique_bm25_hit.chunk_id


@pytest.mark.asyncio
async def test_execute_builds_prompt_with_context() -> None:
    user_id = uuid.uuid4()
    conversation = _build_conversation(user_id=user_id)
    user_message = _build_user_message(conversation_id=conversation.id)
    ai_message = _build_ai_message(conversation_id=conversation.id)
    search_hit = _build_search_hit()

    conversation_usecase, _ = _make_conversation_usecase(
        conversation=conversation,
        user_message=user_message,
        ai_message=ai_message,
    )

    quota_guard = AsyncMock()

    knowledge_repository = AsyncMock()
    knowledge_repository.search_vector.return_value = [search_hit]
    knowledge_repository.search_bm25.return_value = []

    embedding_port = AsyncMock()
    embedding_port.embed_query.return_value = [0.1, 0.2, 0.3]

    llm_port = AsyncMock()
    llm_port.generate.return_value = "Answer"

    usecase = AskAIUseCase(
        conversation=conversation_usecase,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
    )

    command = _build_command(user_id=user_id, conversation_id=conversation.id)
    await usecase.execute(command)

    llm_port.generate.assert_awaited_once()
    call_args = llm_port.generate.call_args
    prompt_text = call_args.args[0]
    assert "Context:" in prompt_text
    assert search_hit.chunk_text in prompt_text
    assert command.prompt in prompt_text
