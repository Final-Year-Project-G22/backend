from __future__ import annotations

import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

import grpc
from core.domain.enums import DocumentSource, Language
from core.domain.exceptions import QuotaExceededError
from core.domain.stream_events import AskStreamEvent, AskStreamEventType
from core.domain.value_objects import SearchHit
from infrastructure.rpc.services.inference_service import AIInferenceService


def _make_request(**overrides: object) -> SimpleNamespace:
    payload: dict[str, object] = {
        "request_id": str(uuid.uuid4()),
        "user_id": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
        "query": "What are the key clauses?",
        "language": "en",
        "session_id": "",
        "title": "",
        "top_k": 5,
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


def _mock_usecase(
    *,
    stream_chunks: tuple[str, ...],
    hits: list[SearchHit],
    usage: SimpleNamespace | None = None,
) -> AsyncMock:
    usecase = AsyncMock()
    conversation = SimpleNamespace(
        id=uuid.uuid4(),
        created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T10:00:00Z"),
        updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T11:00:00Z"),
    )
    user_message = SimpleNamespace(token_usage=None)
    ai_message = SimpleNamespace(token_usage=usage, id=uuid.uuid4(), cache_hit=False)

    usecase._resolve_conversation = AsyncMock(return_value=conversation)
    usecase._persist_user_message = AsyncMock(return_value=user_message)
    usecase._update_user_message_embedding = AsyncMock(return_value=user_message)
    usecase._retrieve_context = AsyncMock(return_value=(hits, []))
    usecase._merge_and_dedupe_hits = lambda vector_hits, _bm25_hits: vector_hits
    usecase._build_prompt = lambda _p, _h: "prompt"
    usecase._persist_ai_message = AsyncMock(return_value=ai_message)
    usecase._cache_response = AsyncMock()
    usecase._publish_query_event = AsyncMock()
    usecase._build_cache_key = lambda _cmd, _id: "cache-key"

    embedding_port = AsyncMock()
    embedding_port.embed_query = AsyncMock(return_value=[0.1, 0.2, 0.3])
    usecase._embedding_port = embedding_port

    llm_port = AsyncMock()
    llm_port.model = "test-model"
    usecase._llm_port = llm_port

    async def _mock_stream(*args: object, **kwargs: object):
        for chunk_text in stream_chunks:
            yield AskStreamEvent(type_=AskStreamEventType.TEXT, text=chunk_text)

        yield AskStreamEvent(
            type_=AskStreamEventType.DONE,
            done=conversation,
            ai_message=ai_message,
            merged_hits=hits,
        )

    usecase.execute_stream_with_tools = _mock_stream

    return usecase


@pytest.mark.asyncio
async def test_ask_stream_persists_ai_message_after_stream_with_full_response() -> None:
    usecase = _mock_usecase(stream_chunks=("Hello ", "world"), hits=[])
    service = AIInferenceService(ask_ai_usecase=usecase)

    chunks = [
        chunk
        async for chunk in service.AskStream(_make_request(), SimpleNamespace(abort=AsyncMock()))
    ]

    assert chunks[0].text.text == "Hello "
    assert chunks[1].text.text == "world"


@pytest.mark.asyncio
async def test_ask_stream_includes_citations_at_end() -> None:
    doc_id = uuid.uuid4()
    chunk_id = uuid.uuid4()
    hit = SearchHit(
        document_id=doc_id,
        chunk_id=chunk_id,
        score=0.95,
        chunk_index=0,
        chunk_text="Relevant clause",
        source=DocumentSource.GUIDE,
        language=Language.ENGLISH,
        document_title="Test Guide",
    )

    usecase = _mock_usecase(stream_chunks=("Answer",), hits=[hit])
    service = AIInferenceService(ask_ai_usecase=usecase)

    chunks = [
        chunk
        async for chunk in service.AskStream(_make_request(), SimpleNamespace(abort=AsyncMock()))
    ]

    citations_chunk = chunks[1]
    assert citations_chunk.HasField("citations")
    assert len(citations_chunk.citations.citations) == 1


@pytest.mark.asyncio
async def test_ask_stream_includes_usage_in_done() -> None:
    usage = SimpleNamespace(prompt_tokens=10, completion_tokens=20, total_tokens=30)
    usecase = _mock_usecase(stream_chunks=("Test",), hits=[], usage=usage)
    service = AIInferenceService(ask_ai_usecase=usecase)

    chunks = [
        chunk
        async for chunk in service.AskStream(_make_request(), SimpleNamespace(abort=AsyncMock()))
    ]

    done_chunk = chunks[-1]
    assert done_chunk.done.model == "test-model"
    assert done_chunk.done.usage.total_tokens == 30
    assert done_chunk.done.session_created_at == "2026-01-01T10:00:00Z"
    assert done_chunk.done.session_updated_at == "2026-01-01T11:00:00Z"


@pytest.mark.asyncio
async def test_ask_stream_error_on_quota_exceeded() -> None:
    usecase = AsyncMock()

    async def _raise_quota(*args: object, **kwargs: object):
        raise QuotaExceededError("Limit reached")
        yield  # make this an async generator

    usecase.execute_stream_with_tools = _raise_quota

    service = AIInferenceService(ask_ai_usecase=usecase)
    chunks = [
        chunk
        async for chunk in service.AskStream(_make_request(), SimpleNamespace(abort=AsyncMock()))
    ]

    error_chunk = chunks[0]
    assert error_chunk.HasField("error")
    assert error_chunk.error.code == "RESOURCE_EXHAUSTED"


@pytest.mark.asyncio
async def test_ask_stream_maps_tool_suppressed_when_debug_enabled() -> None:
    usecase = AsyncMock()
    conversation = SimpleNamespace(
        id=uuid.uuid4(),
        created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T10:00:00Z"),
        updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T11:00:00Z"),
    )
    ai_message = SimpleNamespace(token_usage=None, id=uuid.uuid4(), cache_hit=False)
    usecase._llm_port = SimpleNamespace(model="test-model")

    async def _mock_stream(*args: object, **kwargs: object):
        yield AskStreamEvent(
            type_=AskStreamEventType.TOOL_SUPPRESSED,
            tool_name="search_knowledge_base",
            suppression_reason="duplicate_of_prior_search",
            matched_query="Ethiopian government business registration process",
        )
        yield AskStreamEvent(
            type_=AskStreamEventType.DONE,
            done=conversation,
            ai_message=ai_message,
            merged_hits=[],
        )

    usecase.execute_stream_with_tools = _mock_stream
    service = AIInferenceService(ask_ai_usecase=usecase)

    chunks = [
        chunk
        async for chunk in service.AskStream(
            _make_request(debug_mode=True), SimpleNamespace(abort=AsyncMock())
        )
    ]

    suppressed_chunks = [chunk for chunk in chunks if chunk.HasField("tool_suppressed")]
    assert len(suppressed_chunks) == 1
    assert suppressed_chunks[0].tool_suppressed.tool == "search_knowledge_base"
    assert suppressed_chunks[0].tool_suppressed.reason == "duplicate_of_prior_search"
    assert (
        suppressed_chunks[0].tool_suppressed.matched_query
        == "Ethiopian government business registration process"
    )


@pytest.mark.asyncio
async def test_ask_stream_drops_tool_suppressed_when_debug_disabled() -> None:
    usecase = AsyncMock()
    conversation = SimpleNamespace(
        id=uuid.uuid4(),
        created_at=SimpleNamespace(isoformat=lambda: "2026-01-01T10:00:00Z"),
        updated_at=SimpleNamespace(isoformat=lambda: "2026-01-01T11:00:00Z"),
    )
    ai_message = SimpleNamespace(token_usage=None, id=uuid.uuid4(), cache_hit=False)
    usecase._llm_port = SimpleNamespace(model="test-model")

    async def _mock_stream(*args: object, **kwargs: object):
        yield AskStreamEvent(
            type_=AskStreamEventType.TOOL_SUPPRESSED,
            tool_name="search_knowledge_base",
            suppression_reason="drift",
            matched_query="original prompt",
        )
        yield AskStreamEvent(
            type_=AskStreamEventType.DONE,
            done=conversation,
            ai_message=ai_message,
            merged_hits=[],
        )

    usecase.execute_stream_with_tools = _mock_stream
    service = AIInferenceService(ask_ai_usecase=usecase)

    chunks = [
        chunk
        async for chunk in service.AskStream(_make_request(), SimpleNamespace(abort=AsyncMock()))
    ]

    assert all(not chunk.HasField("tool_suppressed") for chunk in chunks)


@pytest.mark.asyncio
async def test_ask_stream_aborts_when_feature_flag_disabled() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase, ask_enabled=False)
    context = SimpleNamespace(abort=AsyncMock())

    chunks = [chunk async for chunk in service.AskStream(_make_request(), context)]

    assert chunks == []
    context.abort.assert_awaited_once_with(grpc.StatusCode.UNAVAILABLE, "Ask API is disabled")
