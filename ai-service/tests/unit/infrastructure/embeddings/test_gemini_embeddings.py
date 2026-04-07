from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import EmbeddingError
from infrastructure.embeddings.gemini import GeminiEmbeddingAdapter


@pytest.mark.asyncio
async def test_gemini_embed_query_returns_vector() -> None:
    payload = {"embedding": {"values": [0.5, 0.6, 0.7]}}

    async def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.params.get("key") == "test-key"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiEmbeddingAdapter(
            api_key="test-key",
            model="text-embedding-004",
            dimensions=3,
            http_client=client,
        )

        result = await adapter.embed_query("እንዴት እመዝግባ?")

    expected_dimensions = 3
    assert result == [0.5, 0.6, 0.7]
    assert adapter.provider == "gemini"
    assert adapter.dimensions == expected_dimensions


@pytest.mark.asyncio
async def test_gemini_embed_documents_returns_vectors() -> None:
    payload = {"embedding": {"values": [0.1, 0.2]}}
    captured_payload: dict[str, object] = {}

    async def handler(request: httpx.Request) -> httpx.Response:
        captured_payload.update(json.loads(request.content))
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiEmbeddingAdapter(
            api_key="test-key",
            model="gemini-embedding-2-preview",
            dimensions=2,
            http_client=client,
        )

        result = await adapter.embed_documents(
            ["doc one", "doc two"],
            input_type="search_document",
        )

    assert result == [[0.1, 0.2], [0.1, 0.2]]
    assert captured_payload.get("taskType") == "RETRIEVAL_DOCUMENT"


@pytest.mark.asyncio
async def test_gemini_embed_handles_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(400, json={"error": {"message": "bad"}})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = GeminiEmbeddingAdapter(
            api_key="test-key",
            model="text-embedding-004",
            dimensions=3,
            http_client=client,
        )

        with pytest.raises(EmbeddingError, match="gemini embedding request failed"):
            await adapter.embed_query("test")
