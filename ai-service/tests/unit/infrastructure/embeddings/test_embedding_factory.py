from __future__ import annotations

import httpx
import pytest

from app.config import Settings
from core.domain.exceptions import ConfigurationError
from infrastructure.embeddings import (
    CohereEmbeddingAdapter,
    GeminiEmbeddingAdapter,
    OllamaEmbeddingAdapter,
    create_embedding_adapter,
)


@pytest.mark.asyncio
async def test_embedding_factory_returns_cohere_adapter() -> None:
    settings = Settings(
        EMBEDDING_PROVIDER="cohere",
        COHERE_API_KEY="test-key",
        COHERE_EMBEDDING_MODEL="embed-multilingual-v3.0",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_embedding_adapter(settings, http_client=client)

    assert isinstance(adapter, CohereEmbeddingAdapter)


@pytest.mark.asyncio
async def test_embedding_factory_returns_gemini_adapter() -> None:
    settings = Settings(
        EMBEDDING_PROVIDER="gemini",
        GEMINI_API_KEY="test-key",
        GEMINI_EMBEDDING_MODEL="gemini-embedding-2-preview",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_embedding_adapter(settings, http_client=client)

    assert isinstance(adapter, GeminiEmbeddingAdapter)


@pytest.mark.asyncio
async def test_embedding_factory_returns_ollama_adapter() -> None:
    settings = Settings(
        EMBEDDING_PROVIDER="ollama",
        OLLAMA_BASE_URL="http://localhost:11434",
        OLLAMA_EMBEDDING_MODEL="nomic-embed-text",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_embedding_adapter(settings, http_client=client)

    assert isinstance(adapter, OllamaEmbeddingAdapter)


@pytest.mark.asyncio
async def test_embedding_factory_raises_for_unknown_provider() -> None:
    settings = Settings(EMBEDDING_PROVIDER="unknown")
    async with httpx.AsyncClient() as client:
        with pytest.raises(ConfigurationError, match="unsupported embedding provider"):
            create_embedding_adapter(settings, http_client=client)
