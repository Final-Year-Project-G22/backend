from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk
from infrastructure.llm.ollama import OllamaLLMAdapter

MAX_TOKENS = 64
TEMPERATURE = 0.6


@pytest.mark.asyncio
async def test_ollama_generate_returns_text() -> None:
    payload = {"message": {"content": "መልስ"}}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["model"] == "qwen2.5"
        assert body["messages"][0] == {"role": "user", "content": "Hello"}
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        result = await adapter.generate("Hello")

    assert result.text == "መልስ"
    assert result.tool_calls is None
    assert adapter.provider == "ollama"


@pytest.mark.asyncio
async def test_ollama_generate_sends_expected_payload_fields() -> None:
    payload = {"message": {"content": "ok"}}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["model"] == "qwen2.5"
        assert body["messages"][0] == {"role": "user", "content": "Hello"}
        assert body["stream"] is False
        assert body["options"]["num_predict"] == MAX_TOKENS
        assert body["options"]["temperature"] == TEMPERATURE
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        await adapter.generate("Hello", max_tokens=MAX_TOKENS, temperature=TEMPERATURE)


@pytest.mark.asyncio
async def test_ollama_generate_stream_yields_chunks() -> None:
    lines = [
        json.dumps({"message": {"content": "Hel"}, "done": False}),
        json.dumps({"message": {"content": "lo"}, "done": True}),
    ]

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text="\n".join(lines))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        chunks = [chunk async for chunk in adapter.generate_stream("Hello")]

    assert chunks == [LLMChunk(text="Hel"), LLMChunk(text="lo")]


@pytest.mark.asyncio
async def test_ollama_generate_stream_ignores_invalid_json_chunks() -> None:
    lines = [
        "not-json",
        json.dumps({"message": {"content": "ok"}, "done": True}),
    ]

    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text="\n".join(lines))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        chunks = [chunk async for chunk in adapter.generate_stream("Hello")]

    assert chunks == [LLMChunk(text="ok")]


@pytest.mark.asyncio
async def test_ollama_generate_handles_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        with pytest.raises(LLMError, match="ollama generation failed"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_ollama_generate_raises_on_missing_response() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        with pytest.raises(LLMError, match="missing message"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_ollama_generate_raises_on_non_string_response() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"message": {"content": 123}})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        with pytest.raises(LLMError, match="missing text"):
            await adapter.generate("Hello")


@pytest.mark.asyncio
async def test_ollama_generate_stream_raises_on_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaLLMAdapter(
            base_url="http://localhost:11434",
            model="qwen2.5",
            http_client=client,
        )

        with pytest.raises(LLMError, match="streaming generation failed"):
            async for _ in adapter.generate_stream("Hello"):
                pass
