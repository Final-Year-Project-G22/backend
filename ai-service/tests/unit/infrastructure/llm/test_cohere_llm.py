from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import LLMError
from infrastructure.llm.cohere import CohereLLMAdapter


@pytest.mark.asyncio
async def test_cohere_generate_returns_text() -> None:
    payload = {"message": {"content": "እሺ"}}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["model"] == "command-r"
        assert body["message"] == "Hello"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereLLMAdapter(
            api_key="test-key",
            model="command-r",
            http_client=client,
        )

        result = await adapter.generate("Hello")

    assert result == "እሺ"
    assert adapter.provider == "cohere"


@pytest.mark.asyncio
async def test_cohere_generate_stream_yields_chunks() -> None:
    lines = [
        'data: {"message": {"content": "Hel"}}',
        'data: {"message": {"content": "lo"}}',
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

    assert chunks == ["Hel", "lo"]


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
