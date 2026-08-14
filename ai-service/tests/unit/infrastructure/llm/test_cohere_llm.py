from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk
from infrastructure.llm.cohere import CohereLLMAdapter

MAX_TOKENS = 128
TEMPERATURE = 0.5


@pytest.mark.asyncio
async def test_cohere_generate_returns_text() -> None:
    payload = {"message": {"content": [{"type": "text", "text": "እሺ"}]}}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["model"] == "command-r"
        assert body["messages"][0] == {"role": "user", "content": "Hello"}
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        result = await adapter.generate("Hello")

    assert result.text == "እሺ"
    assert result.tool_calls is None
    assert adapter.provider == "cohere"


@pytest.mark.asyncio
async def test_cohere_generate_sends_expected_payload_fields() -> None:
    payload = {"message": {"content": [{"type": "text", "text": "ok"}]}}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["model"] == "command-r"
        assert body["messages"][0] == {"role": "user", "content": "Hello"}
        assert body["max_tokens"] == MAX_TOKENS
        assert body["temperature"] == TEMPERATURE
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        await adapter.generate("Hello", max_tokens=MAX_TOKENS, temperature=TEMPERATURE)


@pytest.mark.asyncio
async def test_cohere_generate_stream_yields_chunks() -> None:
    lines = [
        'data: {"type": "content-delta", "delta": {"message": {"content": {"text": "Hel"}}}}',
        'data: {"type": "content-delta", "delta": {"message": {"content": {"text": "lo"}}}}',
        "data: [DONE]",
    ]

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text="\n".join(lines))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        chunks = [chunk async for chunk in adapter.generate_stream("Hello")]

    assert chunks == [LLMChunk(text="Hel"), LLMChunk(text="lo")]


@pytest.mark.asyncio
async def test_cohere_generate_stream_ignores_invalid_json_chunks() -> None:
    lines = [
        "data: not-json",
        'data: {"type": "content-delta", "delta": {"message": {"content": {"text": "ok"}}}}',
        "data: [DONE]",
    ]

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text="\n".join(lines))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        chunks = [chunk async for chunk in adapter.generate_stream("Hello")]

    assert chunks == [LLMChunk(text="ok")]


@pytest.mark.asyncio
async def test_cohere_generate_handles_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        with pytest.raises(LLMError, match="cohere generation failed"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_cohere_generate_raises_on_missing_content() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"message": {}})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        with pytest.raises(LLMError, match="missing content"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_cohere_generate_stream_raises_on_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        with pytest.raises(LLMError, match="streaming generation failed"):
            async for _ in adapter.generate_stream("Hello"):
                pass
