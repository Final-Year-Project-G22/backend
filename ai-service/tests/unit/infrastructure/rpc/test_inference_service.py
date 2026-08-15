from __future__ import annotations

import uuid
from collections.abc import AsyncIterator
from types import SimpleNamespace
from unittest.mock import AsyncMock

import grpc
import pytest

from core.domain.enums import DocumentSource, Language
from core.domain.exceptions import AIServiceError, QuotaExceededError, RepositoryError
from core.domain.stream_events import AskStreamEvent, AskStreamEventType
from core.domain.value_objects import SearchHit
from core.usecases.contracts import AskAICommand
from core.usecases.defaults import MAX_PROMPT_LENGTH
from infrastructure.rpc.services.inference_service import AIInferenceService

TOP_K_INPUT = 7
PROMPT_TOKENS = 11
COMPLETION_TOKENS = 22
TOTAL_TOKENS = 33


def _make_request(**overrides: object) -> SimpleNamespace:
    payload: dict[str, object] = {
        "request_id": str(uuid.uuid4()),
        "user_id": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
        "query": "How do I register a business?",
        "language": "en",
        "session_id": "",
        "title": "",
        "top_k": 5,
        "strategy": "simple",
        "debug_mode": False,
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


@pytest.mark.asyncio
async def test_ask_maps_successful_response() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase)

    convo_id = uuid.uuid4()
    hit_document_id = uuid.uuid4()
    hit_chunk_id = uuid.uuid4()
    usecase.execute.return_value = SimpleNamespace(
        conversation=SimpleNamespace(
            id=convo_id,
            created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T10:00:00Z"),
            updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T11:00:00Z"),
        ),
        ai_message=SimpleNamespace(
            llm_response="Start by obtaining your trade license.",
            token_usage=SimpleNamespace(
                prompt_tokens=PROMPT_TOKENS,
                completion_tokens=COMPLETION_TOKENS,
                total_tokens=TOTAL_TOKENS,
            ),
        ),
        retrieved_hits=[
            SearchHit(
                document_id=hit_document_id,
                chunk_id=hit_chunk_id,
                score=0.87,
                chunk_text="Relevant text",
                chunk_index=0,
                source=DocumentSource.GUIDE,
                language=Language.ENGLISH,
            )
        ],
    )

    request = _make_request(language="am", top_k=TOP_K_INPUT, title="Custom Conversation")
    context = AsyncMock()

    response = await service.Ask(request, context)

    usecase.execute.assert_awaited_once()
    command = usecase.execute.await_args.args[0]
    assert isinstance(command, AskAICommand)
    assert command.user_id == uuid.UUID(request.user_id)
    assert command.prompt == request.query
    assert command.title == "Custom Conversation"
    assert command.vector_top_k == TOP_K_INPUT

    assert response.request_id == request.request_id
    assert response.session_id == str(convo_id)
    assert response.session_created_at == "2026-01-01T10:00:00Z"
    assert response.session_updated_at == "2026-01-01T11:00:00Z"
    assert response.answer == "Start by obtaining your trade license."
    assert len(response.citations) == 1
    assert response.citations[0].document_id == str(hit_document_id)
    assert response.citations[0].chunk_id == str(hit_chunk_id)
    assert response.usage.prompt_tokens == PROMPT_TOKENS
    assert response.usage.total_tokens == TOTAL_TOKENS


@pytest.mark.asyncio
async def test_ask_aborts_for_prompt_validation() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request(query="x" * (MAX_PROMPT_LENGTH + 1))
    context = AsyncMock()

    await service.Ask(request, context)

    context.abort.assert_awaited_once()
    assert context.abort.await_args.args[0] is grpc.StatusCode.INVALID_ARGUMENT
    assert "prompt" in context.abort.await_args.args[1].lower()
    usecase.execute.assert_not_awaited()


@pytest.mark.asyncio
async def test_ask_aborts_for_invalid_uuid() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request(user_id="not-a-uuid")
    context = AsyncMock()
    context.abort.side_effect = grpc.aio.AbortError(
        grpc.StatusCode.INVALID_ARGUMENT,
        "Invalid UUID or enum: badly formed hexadecimal UUID string",
    )

    with pytest.raises(grpc.aio.AbortError):
        await service.Ask(request, context)

    context.abort.assert_awaited_once()
    assert context.abort.await_args.args[0] is grpc.StatusCode.INVALID_ARGUMENT
    usecase.execute.assert_not_awaited()


@pytest.mark.asyncio
@pytest.mark.parametrize(
    ("error", "status_code"),
    [
        (QuotaExceededError("quota reached"), grpc.StatusCode.RESOURCE_EXHAUSTED),
        (RepositoryError("db down"), grpc.StatusCode.INTERNAL),
        (AIServiceError("provider failed"), grpc.StatusCode.INTERNAL),
    ],
)
async def test_ask_maps_domain_errors_to_grpc_status(
    error: Exception, status_code: grpc.StatusCode
) -> None:
    usecase = AsyncMock()
    usecase.execute.side_effect = error
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request()
    context = AsyncMock()

    await service.Ask(request, context)

    context.abort.assert_awaited_once()
    assert context.abort.await_args.args[0] is status_code


@pytest.mark.asyncio
async def test_ask_aborts_when_feature_flag_disabled() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase, ask_enabled=False)

    request = _make_request()
    context = AsyncMock()
    context.abort.side_effect = grpc.aio.AbortError(
        grpc.StatusCode.UNAVAILABLE,
        "Ask API is disabled",
    )

    with pytest.raises(grpc.aio.AbortError):
        await service.Ask(request, context)

    context.abort.assert_awaited_once_with(
        grpc.StatusCode.UNAVAILABLE,
        "Ask API is disabled",
    )
    usecase.execute.assert_not_awaited()


def _make_stream_usecase(events: list[AskStreamEvent]) -> SimpleNamespace:
    async def stream(_command: AskAICommand) -> AsyncIterator[AskStreamEvent]:
        for event in events:
            yield event

    return SimpleNamespace(execute_stream_with_tools=stream, llm_model="test-model")


def _make_done_event() -> tuple[AskStreamEvent, uuid.UUID]:
    convo_id = uuid.uuid4()
    hit = SearchHit(
        document_id=uuid.uuid4(),
        chunk_id=uuid.uuid4(),
        score=0.87,
        chunk_text="Relevant text",
        chunk_index=0,
        source=DocumentSource.GUIDE,
        language=Language.ENGLISH,
    )
    event = AskStreamEvent(
        type_=AskStreamEventType.DONE,
        done=SimpleNamespace(
            id=convo_id,
            created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T10:00:00Z"),
            updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T11:00:00Z"),
        ),
        ai_message=SimpleNamespace(
            token_usage=SimpleNamespace(
                prompt_tokens=PROMPT_TOKENS,
                completion_tokens=COMPLETION_TOKENS,
                total_tokens=TOTAL_TOKENS,
            )
        ),
        merged_hits=[hit],
    )
    return event, convo_id


@pytest.mark.asyncio
async def test_ask_stream_maps_each_event_type_to_its_chunk() -> None:
    events = [
        AskStreamEvent(type_=AskStreamEventType.TEXT, text="Hello there"),
        AskStreamEvent(
            type_=AskStreamEventType.TOOL_CALL,
            tool_name="search_knowledge_base",
            tool_arguments={"query": "trade licence"},
        ),
        AskStreamEvent(
            type_=AskStreamEventType.TOOL_RESULT,
            tool_name="search_knowledge_base",
            tool_result_summary="3 hits",
        ),
        AskStreamEvent(type_=AskStreamEventType.THINKING, text="Thinking out loud"),
        AskStreamEvent(
            type_=AskStreamEventType.TOOL_SUPPRESSED,
            tool_name="search_knowledge_base",
            suppression_reason="duplicate_of_prior_search",
            matched_query="trade licence",
        ),
    ]
    usecase = _make_stream_usecase(events)
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request(debug_mode=True)
    context = AsyncMock()

    chunks = [chunk async for chunk in service.AskStream(request, context)]

    assert [chunk.WhichOneof("chunk") for chunk in chunks] == [
        "text",
        "tool_use",
        "tool_result",
        "thinking",
        "tool_suppressed",
    ]
    assert chunks[0].text.text == "Hello there"
    assert chunks[1].tool_use.tool == "search_knowledge_base"
    assert chunks[1].tool_use.arguments_json == '{"query": "trade licence"}'
    assert chunks[2].tool_result.tool == "search_knowledge_base"
    assert chunks[2].tool_result.result_summary == "3 hits"
    assert chunks[3].thinking.text == "Thinking out loud"
    assert chunks[4].tool_suppressed.tool == "search_knowledge_base"
    assert chunks[4].tool_suppressed.reason == "duplicate_of_prior_search"
    assert chunks[4].tool_suppressed.matched_query == "trade licence"
    context.abort.assert_not_awaited()


@pytest.mark.asyncio
async def test_ask_stream_filters_debug_chunks_without_debug_mode() -> None:
    events = [
        AskStreamEvent(type_=AskStreamEventType.TEXT, text="Answer"),
        AskStreamEvent(type_=AskStreamEventType.THINKING, text="Hidden thinking"),
        AskStreamEvent(
            type_=AskStreamEventType.TOOL_SUPPRESSED,
            tool_name="search_knowledge_base",
            suppression_reason="duplicate_of_prior_search",
            matched_query="q",
        ),
    ]
    usecase = _make_stream_usecase(events)
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request(debug_mode=False)
    context = AsyncMock()

    chunks = [chunk async for chunk in service.AskStream(request, context)]

    assert [chunk.WhichOneof("chunk") for chunk in chunks] == ["text"]


@pytest.mark.asyncio
async def test_ask_stream_emits_citations_then_done_chunk() -> None:
    done_event, convo_id = _make_done_event()
    usecase = _make_stream_usecase(
        [AskStreamEvent(type_=AskStreamEventType.TEXT, text="Answer"), done_event]
    )
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request()
    context = AsyncMock()

    chunks = [chunk async for chunk in service.AskStream(request, context)]

    assert [chunk.WhichOneof("chunk") for chunk in chunks] == ["text", "citations", "done"]
    citations = chunks[1].citations.citations
    assert len(citations) == 1
    assert citations[0].chunk_id == str(done_event.merged_hits[0].chunk_id)
    done = chunks[2].done
    assert done.model == "test-model"
    assert done.session_id == str(convo_id)
    assert done.session_created_at == "2026-01-01T10:00:00Z"
    assert done.session_updated_at == "2026-01-01T11:00:00Z"
    assert done.usage.prompt_tokens == PROMPT_TOKENS
    assert done.usage.completion_tokens == COMPLETION_TOKENS
    assert done.usage.total_tokens == TOTAL_TOKENS
    assert done.latency_ms >= 0


@pytest.mark.asyncio
async def test_ask_stream_stops_after_events_when_no_done_event_arrives() -> None:
    usecase = _make_stream_usecase(
        [AskStreamEvent(type_=AskStreamEventType.TEXT, text="Only text")]
    )
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request()
    context = AsyncMock()

    chunks = [chunk async for chunk in service.AskStream(request, context)]

    assert [chunk.WhichOneof("chunk") for chunk in chunks] == ["text"]


@pytest.mark.asyncio
@pytest.mark.parametrize(
    ("error", "expected_code"),
    [
        (QuotaExceededError("quota reached"), "RESOURCE_EXHAUSTED"),
        (RuntimeError("boom"), "INTERNAL"),
    ],
)
async def test_ask_stream_emits_error_chunk_on_failure(
    error: Exception, expected_code: str
) -> None:
    async def failing_stream(_command: AskAICommand) -> AsyncIterator[AskStreamEvent]:
        if False:  # pragma: no cover
            yield AskStreamEvent(type_=AskStreamEventType.TEXT, text="")
        raise error

    usecase = SimpleNamespace(
        execute_stream_with_tools=failing_stream,
        llm_model="test-model",
    )
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request()
    context = AsyncMock()

    chunks = [chunk async for chunk in service.AskStream(request, context)]

    assert len(chunks) == 1
    assert chunks[0].WhichOneof("chunk") == "error"
    assert chunks[0].error.code == expected_code
