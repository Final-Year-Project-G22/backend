from __future__ import annotations

import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

import grpc
from core.domain.exceptions import QuotaExceededError
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


async def _fake_stream(*chunks: str):
    for chunk in chunks:
        yield chunk


def _mock_usecase(
    *,
    stream_chunks: tuple[str, ...],
    hits: list[SimpleNamespace],
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
    llm_port.generate_stream = lambda _prompt: _fake_stream(*stream_chunks)
    usecase._llm_port = llm_port

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
    usecase._persist_ai_message.assert_awaited_once()
    persisted_answer = usecase._persist_ai_message.await_args.args[2]
    assert persisted_answer == "Hello world"


@pytest.mark.asyncio
async def test_ask_stream_includes_citations_at_end() -> None:
    doc_id = uuid.uuid4()
    chunk_id = uuid.uuid4()
    hit = SimpleNamespace(document_id=doc_id, chunk_id=chunk_id, score=0.95)

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
    usecase._resolve_conversation = AsyncMock(side_effect=QuotaExceededError("Limit reached"))

    service = AIInferenceService(ask_ai_usecase=usecase)
    chunks = [
        chunk
        async for chunk in service.AskStream(_make_request(), SimpleNamespace(abort=AsyncMock()))
    ]

    error_chunk = chunks[0]
    assert error_chunk.HasField("error")
    assert error_chunk.error.code == "RESOURCE_EXHAUSTED"


@pytest.mark.asyncio
async def test_ask_stream_aborts_when_feature_flag_disabled() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase, ask_enabled=False)
    context = SimpleNamespace(abort=AsyncMock())

    chunks = [chunk async for chunk in service.AskStream(_make_request(), context)]

    assert chunks == []
    context.abort.assert_awaited_once_with(grpc.StatusCode.UNAVAILABLE, "Ask API is disabled")
