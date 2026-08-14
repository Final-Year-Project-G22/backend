from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk
from infrastructure.llm.gemini import GeminiLLMAdapter

MAX_TOKENS = 128
TEMPERATURE = 0.5


@pytest.mark.asyncio
async def test_gemini_generate_returns_text() -> None:
    payload = {"candidates": [{"content": {"parts": [{"text": "መልስ"}]}}]}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["contents"][0]["parts"][0]["text"] == "Hello"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        result = await adapter.generate("Hello")

    assert result.text == "መልስ"
    assert result.tool_calls is None
    assert adapter.provider == "gemini"


@pytest.mark.asyncio
async def test_gemini_generate_sends_expected_payload_fields() -> None:
    payload = {"candidates": [{"content": {"parts": [{"text": "ok"}]}}]}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        config = body["generationConfig"]
        assert body["contents"][0]["parts"][0]["text"] == "Hello"
        assert config["maxOutputTokens"] == MAX_TOKENS
        assert config["temperature"] == TEMPERATURE
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        await adapter.generate("Hello", max_tokens=MAX_TOKENS, temperature=TEMPERATURE)


@pytest.mark.asyncio
async def test_gemini_generate_stream_yields_chunks() -> None:
    lines = [
        json.dumps({"candidates": [{"content": {"parts": [{"text": "Hel"}]}}]}),
        json.dumps({"candidates": [{"content": {"parts": [{"text": "lo"}]}}]}),
    ]

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text="\n".join(lines))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        chunks = [chunk async for chunk in adapter.generate_stream("Hello")]

    assert chunks == [LLMChunk(text="Hel"), LLMChunk(text="lo")]


@pytest.mark.asyncio
async def test_gemini_generate_stream_ignores_invalid_json_chunks() -> None:
    lines = [
        "not-json",
        json.dumps({"candidates": [{"content": {"parts": [{"text": "ok"}]}}]}),
    ]

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text="\n".join(lines))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        chunks = [chunk async for chunk in adapter.generate_stream("Hello")]

    assert chunks == [LLMChunk(text="ok")]


@pytest.mark.asyncio
async def test_gemini_generate_handles_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(400, json={"error": {"message": "bad"}})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        with pytest.raises(LLMError, match="gemini generation failed"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_gemini_generate_raises_on_missing_candidates() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        with pytest.raises(LLMError, match="missing text"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_gemini_generate_raises_on_non_string_text() -> None:
    payload = {"candidates": [{"content": {"parts": [{"text": 123}]}}]}

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        with pytest.raises(LLMError, match="missing text"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_gemini_generate_stream_raises_on_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiLLMAdapter(
            api_key="test-key",
            model="gemini-1.5-flash",
            http_client=client,
        )

        with pytest.raises(LLMError, match="streaming generation failed"):
            async for _ in adapter.generate_stream("Hello"):
                pass
