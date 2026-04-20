from __future__ import annotations

import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

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
        "top_k": 5,
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


async def _fake_stream(*chunks: str):
    for chunk in chunks:
        yield chunk


@pytest.mark.asyncio
async def test_ask_stream_yields_text_chunks() -> None:
    usecase = AsyncMock()
    response = SimpleNamespace(
        conversation=SimpleNamespace(id=uuid.uuid4()),
        retrieved_hits=[],
        ai_message=SimpleNamespace(token_usage=None),
    )
    usecase.execute = AsyncMock(return_value=response)

    llm_port = AsyncMock()
    llm_port.model = "test-model"
    llm_port.generate_stream = lambda _prompt: _fake_stream("Hello ", "world")
    usecase._llm_port = llm_port
    usecase._build_prompt = lambda _p, _h: "test prompt"

    service = AIInferenceService(ask_ai_usecase=usecase)
    request = _make_request()
    context = SimpleNamespace(abort=AsyncMock())

    chunks = [chunk async for chunk in service.AskStream(request, context)]

    assert len(chunks) >= 2
    assert chunks[0].text.text == "Hello "
    assert chunks[1].text.text == "world"


@pytest.mark.asyncio
async def test_ask_stream_includes_citations_at_end() -> None:
    doc_id = uuid.uuid4()
    chunk_id = uuid.uuid4()

    usecase = AsyncMock()
    response = SimpleNamespace(
        conversation=SimpleNamespace(id=uuid.uuid4()),
        retrieved_hits=[
            SimpleNamespace(
                document_id=doc_id,
                chunk_id=chunk_id,
                score=0.95,
            )
        ],
        ai_message=SimpleNamespace(token_usage=None),
    )
    usecase.execute = AsyncMock(return_value=response)

    llm_port = AsyncMock()
    llm_port.model = "test-model"
    llm_port.generate_stream = lambda _prompt: _fake_stream("Answer")
    usecase._llm_port = llm_port
    usecase._build_prompt = lambda _p, _h: "prompt"

    service = AIInferenceService(ask_ai_usecase=usecase)
    request = _make_request()

    context = SimpleNamespace(abort=AsyncMock())
    chunks = [chunk async for chunk in service.AskStream(request, context)]

    citations_chunk = chunks[1]
    assert citations_chunk.HasField("citations")
    assert len(citations_chunk.citations.citations) == 1


@pytest.mark.asyncio
async def test_ask_stream_includes_usage_in_done() -> None:
    usecase = AsyncMock()
    response = SimpleNamespace(
        conversation=SimpleNamespace(id=uuid.uuid4()),
        retrieved_hits=[],
        ai_message=SimpleNamespace(
            token_usage=SimpleNamespace(
                prompt_tokens=10,
                completion_tokens=20,
                total_tokens=30,
            )
        ),
    )
    usecase.execute = AsyncMock(return_value=response)

    llm_port = AsyncMock()
    llm_port.model = "test-model"
    llm_port.generate_stream = lambda _prompt: _fake_stream("Test")
    usecase._llm_port = llm_port
    usecase._build_prompt = lambda _p, _h: "prompt"

    service = AIInferenceService(ask_ai_usecase=usecase)
    request = _make_request()

    context = SimpleNamespace(abort=AsyncMock())
    chunks = [chunk async for chunk in service.AskStream(request, context)]

    done_chunk = chunks[-1]
    assert done_chunk.done.model == "test-model"
    assert done_chunk.done.usage.total_tokens == 30


@pytest.mark.asyncio
async def test_ask_stream_error_on_quota_exceeded() -> None:
    usecase = AsyncMock()
    usecase.execute = AsyncMock(side_effect=QuotaExceededError("Limit reached"))

    service = AIInferenceService(ask_ai_usecase=usecase)
    request = _make_request()

    context = SimpleNamespace(abort=AsyncMock())
    chunks = [chunk async for chunk in service.AskStream(request, context)]

    error_chunk = chunks[0]
    assert error_chunk.HasField("error")
    assert error_chunk.error.code == "RESOURCE_EXHAUSTED"
