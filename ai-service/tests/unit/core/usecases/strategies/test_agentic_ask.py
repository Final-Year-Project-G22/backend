from __future__ import annotations

import uuid
from unittest.mock import AsyncMock, MagicMock

import pytest

from core.domain.enums import DocumentSource, Language, Tier
from core.domain.models import AIConversationSession
from core.domain.stream_events import AskStreamEventType
from core.domain.tools import ToolResult
from core.domain.value_objects import SearchHit
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.embedding import EmbeddingPort
from core.ports.intent_classifier import IntentClass, IntentClassifierPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from core.ports.llm import LLMPort, LLMResult, ToolCall, ToolDefinition
from core.ports.tool_registry import ToolRegistryPort
from core.usecases.contracts import AskAICommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.strategies.agentic_ask import AgenticAskStrategy
from infrastructure.prefetch.pipeline import PreFetchPipeline
from infrastructure.prompts import PromptLoader


def _make_hit(language: Language) -> SearchHit:
    return SearchHit(
        chunk_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        score=0.95,
        chunk_text="Amharic trade licence context",
        chunk_index=0,
        source=DocumentSource.GOVERNMENT,
        language=language,
        document_title="Trade Licence Guide",
    )


def _make_strategy(
    *,
    vector_hits: list[SearchHit],
    bm25_hits: list[SearchHit],
    llm_results: list[LLMResult],
) -> tuple[AgenticAskStrategy, MagicMock, MagicMock, AskAICommand, SearchHit]:
    user_id = uuid.uuid4()
    account_id = uuid.uuid4()
    command = AskAICommand(
        user_id=user_id,
        account_id=account_id,
        prompt="What is required for a trade licence?",
        language=Language.AMHARIC,
        strategy="agentic",
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
    repository.list_messages = AsyncMock(return_value=[])
    repository.add_message = AsyncMock(side_effect=lambda message: message)
    repository.update_message = AsyncMock(side_effect=lambda message: message)
    conversation = ConversationUseCase(repository)

    knowledge_repository = MagicMock(spec=KnowledgeRepositoryPort)
    knowledge_repository.search_vector = AsyncMock(return_value=vector_hits)
    knowledge_repository.search_bm25 = AsyncMock(return_value=bm25_hits)

    embedding_port = MagicMock(spec=EmbeddingPort)
    embedding_port.embed_query = AsyncMock(return_value=[0.1, 0.2, 0.3])

    llm_port = MagicMock(spec=LLMPort)
    llm_port.generate = AsyncMock(side_effect=llm_results)

    tool_registry = MagicMock(spec=ToolRegistryPort)
    tool_registry.get_tool_definitions = AsyncMock(
        return_value=[
            ToolDefinition(
                name="search_knowledge_base",
                description="Search the knowledge base",
                parameter_schema_json="{}",
            ),
            ToolDefinition(
                name="get_user_profile",
                description="Get the user profile",
                parameter_schema_json="{}",
            ),
        ]
    )
    tool_registry.execute_tool = AsyncMock()

    prompt_loader = MagicMock(spec=PromptLoader)
    prompt_loader.render_agentic.return_value = "agentic system prompt"

    intent_classifier = MagicMock(spec=IntentClassifierPort)
    intent_classifier.classify = AsyncMock(return_value=IntentClass.KNOWLEDGE)

    pre_fetch_pipeline = MagicMock(spec=PreFetchPipeline)
    pre_fetch_pipeline.pre_fetch = AsyncMock(return_value={})

    strategy = AgenticAskStrategy(
        conversation=conversation,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
        prompt_loader=prompt_loader,
        tool_registry=tool_registry,
        intent_classifier=intent_classifier,
        pre_fetch_pipeline=pre_fetch_pipeline,
    )
    return strategy, repository, tool_registry, command, vector_hits[0]


@pytest.mark.asyncio
async def test_agentic_context_is_injected_once_and_citations_are_preserved() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, repository, tool_registry, command, expected_hit = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[hit],
        llm_results=[LLMResult(text="Grounded answer")],
    )

    events = [event async for event in strategy.execute_stream(command)]

    first_call = strategy._llm_port.generate.await_args_list[0]
    assert "Amharic trade licence context" in first_call.args[0]
    assert "search_knowledge_base" in {definition.name for definition in first_call.kwargs["tools"]}
    assert events[-1].type is AskStreamEventType.DONE

    ai_message = repository.add_message.await_args_list[-1].args[0]
    assert ai_message.retrieved_chunk_ids == [expected_hit.chunk_id]
    assert ai_message.context_chunks == [expected_hit]
    assert ai_message.response_sources[0].chunk_id == expected_hit.chunk_id
    tool_registry.execute_tool.assert_not_awaited()
    strategy._pre_fetch_pipeline.pre_fetch.assert_awaited_once()
    assert strategy._pre_fetch_pipeline.pre_fetch.await_args.kwargs["include_kb"] is False

    filters = strategy._knowledge_repository.search_vector.await_args.kwargs["filters"]
    assert filters.language is Language.AMHARIC
    assert (
        strategy._knowledge_repository.search_bm25.await_args.kwargs["filters"].language
        is Language.AMHARIC
    )


@pytest.mark.asyncio
async def test_agentic_keeps_kb_tool_for_follow_up_but_skips_first_duplicate() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(
                        name="search_knowledge_base",
                        arguments={"query": "What is required for a trade licence?"},
                    )
                ],
            ),
            LLMResult(text="Follow-up grounded answer"),
        ],
    )

    events = [event async for event in strategy.execute_stream(command)]

    assert all(event.tool_name != "search_knowledge_base" for event in events)
    tool_registry.execute_tool.assert_not_awaited()
    first_tools = strategy._llm_port.generate.await_args_list[0].kwargs["tools"]
    second_tools = strategy._llm_port.generate.await_args_list[1].kwargs["tools"]
    assert "search_knowledge_base" in {definition.name for definition in first_tools}
    assert "search_knowledge_base" in {definition.name for definition in second_tools}


@pytest.mark.asyncio
async def test_agentic_allows_a_distinct_kb_follow_up_query() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(
                        name="search_knowledge_base",
                        arguments={"query": "trade licence renewal deadlines"},
                    )
                ],
            ),
            LLMResult(text="Follow-up grounded answer"),
        ],
    )
    tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base",
        result_text="Additional renewal context",
    )

    events = [event async for event in strategy.execute_stream(command)]

    assert any(
        event.type is AskStreamEventType.TOOL_CALL and event.tool_name == "search_knowledge_base"
        for event in events
    )
    tool_registry.execute_tool.assert_awaited_once_with(
        "search_knowledge_base",
        {"query": "trade licence renewal deadlines"},
        str(command.account_id),
        str(command.user_id),
    )
