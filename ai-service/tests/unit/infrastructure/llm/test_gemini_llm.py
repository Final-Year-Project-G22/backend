from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import LLMError
from infrastructure.llm.gemini import GeminiLLMAdapter


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

    assert result == "መልስ"
    assert adapter.provider == "gemini"


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

    assert chunks == ["Hel", "lo"]


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
