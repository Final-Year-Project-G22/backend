from __future__ import annotations

import uuid
from unittest.mock import AsyncMock, MagicMock

import pytest

from core.domain.enums import DocumentSource, Language, MessageType, Tier
from core.domain.models import AIChatMessage, AIConversationSession, ToolCallRecord
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
    prompt: str | None = None,
    embedding_map: dict[str, list[float]] | None = None,
    max_iterations: int = 5,
    debug_mode: bool = False,
) -> tuple[AgenticAskStrategy, MagicMock, MagicMock, AskAICommand, SearchHit]:
    user_id = uuid.uuid4()
    account_id = uuid.uuid4()
    command = AskAICommand(
        user_id=user_id,
        account_id=account_id,
        prompt=prompt or "What is required for a trade licence?",
        language=Language.AMHARIC,
        strategy="agentic",
        debug_mode=debug_mode,
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

    async def fake_embed(query: str) -> list[float]:
        if embedding_map is not None and query in embedding_map:
            return embedding_map[query]
        return [1.0, 0.0]

    embedding_port.embed_query = AsyncMock(side_effect=fake_embed)

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
        max_iterations=max_iterations,
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
        embedding_map={
            "trade licence renewal deadlines": [0.6, 0.8],
        },
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


BROADENED_QUERY = "Ethiopian government business registration process"


@pytest.mark.asyncio
async def test_agentic_dedupes_case_variants_and_forces_finalization() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                ],
            ),
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(
                        name="search_knowledge_base",
                        arguments={"query": "Ethiopian Government Business Registration Process"},
                    )
                ],
            ),
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(
                        name="search_knowledge_base",
                        arguments={"query": "Ethiopian Government Business Registration Process"},
                    )
                ],
            ),
            LLMResult(text="Final grounded answer"),
        ],
        embedding_map={BROADENED_QUERY: [0.6, 0.8]},
    )
    tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base",
        result_text="Registration context",
    )

    events = [event async for event in strategy.execute_stream(command)]

    tool_call_events = [event for event in events if event.is_tool_call]
    assert len(tool_call_events) == 1
    assert tool_call_events[0].tool_arguments == {"query": BROADENED_QUERY}
    tool_registry.execute_tool.assert_awaited_once()
    assert strategy._llm_port.generate.await_count == 4

    text_events = [event.text for event in events if event.is_text and event.text]
    assert len(text_events) == 1
    assert text_events[0] == "Final grounded answer"
    assert events[-1].type is AskStreamEventType.DONE

    ai_message = repository.add_message.await_args_list[-1].args[0]
    records = ai_message.tool_calls or []
    executed = [tc for tc in records if not tc.suppressed]
    suppressed = [tc for tc in records if tc.suppressed]
    assert len(executed) == 1
    assert len(suppressed) == 2
    assert all(tc.suppression_reason == "duplicate_of_prior_search" for tc in suppressed)


@pytest.mark.asyncio
async def test_agentic_emits_tool_suppressed_only_in_debug_mode() -> None:
    hit = _make_hit(Language.AMHARIC)

    def build(*, debug_mode: bool):
        return _make_strategy(
            vector_hits=[hit],
            bm25_hits=[],
            llm_results=[
                LLMResult(
                    text="",
                    tool_calls=[
                        ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                    ],
                ),
                LLMResult(
                    text="",
                    tool_calls=[
                        ToolCall(
                            name="search_knowledge_base",
                            arguments={
                                "query": "  ETHIOPIAN government   business registration process "
                            },
                        )
                    ],
                ),
                LLMResult(
                    text="",
                    tool_calls=[
                        ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                    ],
                ),
                LLMResult(text="Final grounded answer"),
            ],
            embedding_map={BROADENED_QUERY: [0.6, 0.8]},
            debug_mode=debug_mode,
        )

    debug_strategy, _, debug_tool_registry, debug_command, _ = build(debug_mode=True)
    debug_tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base", result_text="Registration context"
    )
    debug_events = [event async for event in debug_strategy.execute_stream(debug_command)]
    suppressed_events = [
        event for event in debug_events if event.type is AskStreamEventType.TOOL_SUPPRESSED
    ]
    assert len(suppressed_events) == 2
    assert suppressed_events[0].suppression_reason == "duplicate_of_prior_search"
    assert suppressed_events[0].matched_query == BROADENED_QUERY

    normal_strategy, _, normal_tool_registry, normal_command, _ = build(debug_mode=False)
    normal_tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base", result_text="Registration context"
    )
    normal_events = [event async for event in normal_strategy.execute_stream(normal_command)]
    assert all(event.type is not AskStreamEventType.TOOL_SUPPRESSED for event in normal_events)


@pytest.mark.asyncio
async def test_agentic_emits_visible_fallback_when_forced_finalization_has_no_text() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                ],
            ),
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                ],
            ),
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                ],
            ),
            LLMResult(text=""),
        ],
        embedding_map={BROADENED_QUERY: [0.6, 0.8]},
    )
    tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base", result_text="Registration context"
    )

    events = [event async for event in strategy.execute_stream(command)]

    text_events = [event.text for event in events if event.is_text and event.text]
    assert text_events, "expected a visible fallback text event"
    assert events[-1].type is AskStreamEventType.DONE


@pytest.mark.asyncio
async def test_agentic_suppresses_drifted_kb_query() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": "how to bake bread"})
                ],
            ),
            LLMResult(text="Answer about trade licences"),
        ],
        embedding_map={"how to bake bread": [0.0, 1.0]},
    )

    events = [event async for event in strategy.execute_stream(command)]

    tool_registry.execute_tool.assert_not_awaited()
    assert all(event.type is not AskStreamEventType.TOOL_CALL for event in events)
    text_events = [event.text for event in events if event.is_text and event.text]
    assert len(text_events) == 1
    assert text_events[0] == "Answer about trade licences"


@pytest.mark.asyncio
async def test_agentic_tripwire_triggers_mid_batch_despite_executed_other_tool() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                ],
            ),
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY}),
                    ToolCall(
                        name="search_knowledge_base", arguments={"query": "  " + BROADENED_QUERY}
                    ),
                    ToolCall(name="get_user_profile", arguments={}),
                ],
            ),
            LLMResult(text="Final grounded answer"),
        ],
        embedding_map={BROADENED_QUERY: [0.6, 0.8]},
    )
    tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base", result_text="Registration context"
    )

    events = [event async for event in strategy.execute_stream(command)]

    assert tool_registry.execute_tool.await_count == 1
    assert strategy._llm_port.generate.await_count == 3
    assert all(event.tool_name != "get_user_profile" for event in events if event.is_tool_call)
    assert events[-1].type is AskStreamEventType.DONE


@pytest.mark.asyncio
async def test_agentic_exhausted_loop_still_emits_visible_text() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(
                        name="search_knowledge_base", arguments={"query": "first distinct lookup"}
                    )
                ],
            ),
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(
                        name="search_knowledge_base", arguments={"query": "second distinct lookup"}
                    )
                ],
            ),
        ],
        max_iterations=2,
        embedding_map={
            "first distinct lookup": [0.6, 0.8],
            "second distinct lookup": [0.6, -0.8],
        },
    )
    tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base", result_text="Some context"
    )

    events = [event async for event in strategy.execute_stream(command)]

    text_events = [event.text for event in events if event.is_text and event.text]
    assert text_events, "expected a visible fallback text event at loop exhaustion"
    assert events[-1].type is AskStreamEventType.DONE
    assert tool_registry.execute_tool.await_count == 2


@pytest.mark.asyncio
async def test_agentic_merges_broadened_kb_hits_into_citations() -> None:
    hit = _make_hit(Language.AMHARIC)
    broadened_hit = SearchHit(
        chunk_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        score=0.9,
        chunk_text="Business registration process context",
        chunk_index=1,
        source=DocumentSource.GOVERNMENT,
        language=Language.ENGLISH,
        document_title="Registration Guide",
    )
    strategy, _repository, tool_registry, command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[
            LLMResult(
                text="",
                tool_calls=[
                    ToolCall(name="search_knowledge_base", arguments={"query": BROADENED_QUERY})
                ],
            ),
            LLMResult(text="Grounded answer with citations"),
        ],
        embedding_map={BROADENED_QUERY: [0.6, 0.8]},
    )
    tool_registry.execute_tool.return_value = ToolResult(
        tool_name="search_knowledge_base",
        result_text="Registration context",
        hits=[broadened_hit],
    )

    events = [event async for event in strategy.execute_stream(command)]

    done_event = events[-1]
    assert done_event.type is AskStreamEventType.DONE
    chunk_ids = {str(h.chunk_id) for h in (done_event.merged_hits or [])}
    assert str(hit.chunk_id) in chunk_ids
    assert str(broadened_hit.chunk_id) in chunk_ids


@pytest.mark.asyncio
async def test_format_history_skips_suppressed_tool_calls() -> None:
    hit = _make_hit(Language.AMHARIC)
    strategy, _repository, _tool_registry, _command, _ = _make_strategy(
        vector_hits=[hit],
        bm25_hits=[],
        llm_results=[LLMResult(text="Grounded answer")],
    )

    ai_message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Answer",
        tool_calls=[
            ToolCallRecord(
                tool_name="search_knowledge_base",
                arguments={"query": "kept query"},
                result_summary="3 results",
                iteration=1,
            ),
            ToolCallRecord(
                tool_name="search_knowledge_base",
                arguments={"query": "duplicate query"},
                result_summary="suppressed lookup",
                iteration=2,
                suppressed=True,
                suppression_reason="duplicate_of_prior_search",
            ),
        ],
        message_order=1,
    )

    history = strategy._format_history([ai_message])

    assert "3 results" in history
    assert "suppressed lookup" not in history
    assert history.count("search_knowledge_base") == 1
