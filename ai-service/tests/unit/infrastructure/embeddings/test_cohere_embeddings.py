from __future__ import annotations

import json

import httpx
import pytest

from core.domain.exceptions import EmbeddingError
from infrastructure.embeddings.cohere import CohereEmbeddingAdapter


@pytest.mark.asyncio
async def test_cohere_embed_query_returns_vector() -> None:
    payload = {"embeddings": [[0.1, 0.2, 0.3]]}

    async def handler(request: httpx.Request) -> httpx.Response:
        assert request.headers.get("Authorization") == "Bearer test-key"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=3,
            http_client=client,
        )

        result = await adapter.embed_query("እንዴት እመዝግባ?")

    expected_dimensions = 3
    assert result == [0.1, 0.2, 0.3]
    assert adapter.provider == "cohere"
    assert adapter.dimensions == expected_dimensions


@pytest.mark.asyncio
async def test_cohere_embed_query_sends_expected_payload_fields() -> None:
    payload = {"embeddings": [[0.1, 0.2, 0.3]]}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["model"] == "embed-multilingual-v3.0"
        assert body["texts"] == ["How do I register?"]
        assert body["input_type"] == "search_query"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=3,
            http_client=client,
        )

        await adapter.embed_query("How do I register?")


@pytest.mark.asyncio
async def test_cohere_embed_documents_returns_vectors() -> None:
    payload = {"embeddings": [[0.1, 0.2], [0.3, 0.4]]}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["texts"] == ["doc one", "doc two"]
        assert body["input_type"] == "search_document"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=2,
            http_client=client,
        )

        result = await adapter.embed_documents(
            ["doc one", "doc two"],
            input_type="search_document",
        )

    assert result == [[0.1, 0.2], [0.3, 0.4]]


@pytest.mark.asyncio
async def test_cohere_embed_documents_without_input_type_defaults_to_search_document() -> None:
    payload = {"embeddings": [[0.1, 0.2]]}

    async def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        assert body["input_type"] == "search_document"
        return httpx.Response(200, json=payload)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=2,
            http_client=client,
        )

        result = await adapter.embed_documents(["doc one"])

    assert result == [[0.1, 0.2]]


@pytest.mark.asyncio
async def test_cohere_embed_handles_http_error() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "bad"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=3,
            http_client=client,
        )

        with pytest.raises(EmbeddingError, match="cohere embedding request failed"):
            await adapter.embed_query("test")


@pytest.mark.asyncio
async def test_cohere_embed_query_raises_on_missing_embeddings_field() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"foo": "bar"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=3,
            http_client=client,
        )

        with pytest.raises(EmbeddingError, match="missing embeddings"):
            await adapter.embed_query("test")


@pytest.mark.asyncio
async def test_cohere_embed_query_raises_on_non_list_embeddings() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"embeddings": "not-a-list"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=3,
            http_client=client,
        )

        with pytest.raises(EmbeddingError, match="missing embeddings"):
            await adapter.embed_query("test")


@pytest.mark.asyncio
async def test_cohere_embed_query_raises_on_non_numeric_embedding_value() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"embeddings": [["bad", 0.1]]})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        adapter = CohereEmbeddingAdapter(
            api_key="test-key",
            model="embed-multilingual-v3.0",
            dimensions=3,
            http_client=client,
        )

        with pytest.raises(EmbeddingError, match="invalid embedding value"):
            await adapter.embed_query("test")
