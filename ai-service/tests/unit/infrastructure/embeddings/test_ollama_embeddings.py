from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import EmbeddingError
from infrastructure.embeddings.ollama import OllamaEmbeddingAdapter


@pytest.mark.asyncio
async def test_ollama_embed_query_returns_vector() -> None:
    payload = {"embedding": [0.9, 0.8]}

    async def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path.endswith("/api/embeddings")
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaEmbeddingAdapter(
            base_url="http://localhost:11434",
            model="nomic-embed-text",
            dimensions=2,
            http_client=client,
        )

        result = await adapter.embed_query("እንዴት እመዝግባ?")

    expected_dimensions = 2
    assert result == [0.9, 0.8]
    assert adapter.provider == "ollama"
    assert adapter.dimensions == expected_dimensions


@pytest.mark.asyncio
async def test_ollama_embed_documents_returns_vectors() -> None:
    payload = {"embedding": [0.2, 0.1]}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["prompt"] in {"doc one", "doc two"}
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaEmbeddingAdapter(
            base_url="http://localhost:11434",
            model="nomic-embed-text",
            dimensions=2,
            http_client=client,
        )

        result = await adapter.embed_documents(["doc one", "doc two"])

    assert result == [[0.2, 0.1], [0.2, 0.1]]


@pytest.mark.asyncio
async def test_ollama_embed_handles_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = OllamaEmbeddingAdapter(
            base_url="http://localhost:11434",
            model="nomic-embed-text",
            dimensions=2,
            http_client=client,
        )

        with pytest.raises(EmbeddingError, match="ollama embedding request failed"):
            await adapter.embed_query("test")
