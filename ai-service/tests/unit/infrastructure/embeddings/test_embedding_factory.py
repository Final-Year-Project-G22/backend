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

OVERRIDE_DIMENSIONS = 256
DEFAULT_COHERE_DIMENSIONS = 1024
DEFAULT_OTHERS_DIMENSIONS = 768


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
async def test_embedding_factory_normalizes_provider_case_and_whitespace() -> None:
    settings = Settings(
        EMBEDDING_PROVIDER="  GeMini  ",
        GEMINI_API_KEY="test-key",
        GEMINI_EMBEDDING_MODEL="gemini-embedding-2-preview",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_embedding_adapter(settings, http_client=client)

    assert isinstance(adapter, GeminiEmbeddingAdapter)


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
        with pytest.raises(ConfigurationError, match="unsupported embedding provider") as exc_info:
            create_embedding_adapter(settings, http_client=client)

    assert exc_info.value.details["provider"] == "unknown"


@pytest.mark.asyncio
async def test_embedding_factory_applies_dimensions_override_for_all_providers() -> None:
    cohere_settings = Settings(
        EMBEDDING_PROVIDER="cohere",
        COHERE_API_KEY="test-key",
        COHERE_EMBEDDING_MODEL="embed-multilingual-v3.0",
        EMBEDDING_DIMENSIONS=OVERRIDE_DIMENSIONS,
    )
    gemini_settings = Settings(
        EMBEDDING_PROVIDER="gemini",
        GEMINI_API_KEY="test-key",
        GEMINI_EMBEDDING_MODEL="text-embedding-004",
        EMBEDDING_DIMENSIONS=OVERRIDE_DIMENSIONS,
    )
    ollama_settings = Settings(
        EMBEDDING_PROVIDER="ollama",
        OLLAMA_BASE_URL="http://localhost:11434",
        OLLAMA_EMBEDDING_MODEL="nomic-embed-text",
        EMBEDDING_DIMENSIONS=OVERRIDE_DIMENSIONS,
    )

    async with httpx.AsyncClient() as client:
        cohere_adapter = create_embedding_adapter(cohere_settings, http_client=client)
        gemini_adapter = create_embedding_adapter(gemini_settings, http_client=client)
        ollama_adapter = create_embedding_adapter(ollama_settings, http_client=client)

    assert cohere_adapter.dimensions == OVERRIDE_DIMENSIONS
    assert gemini_adapter.dimensions == OVERRIDE_DIMENSIONS
    assert ollama_adapter.dimensions == OVERRIDE_DIMENSIONS


@pytest.mark.asyncio
async def test_embedding_factory_uses_default_dimensions_when_unset() -> None:
    cohere_settings = Settings(
        EMBEDDING_PROVIDER="cohere",
        COHERE_API_KEY="test-key",
        COHERE_EMBEDDING_MODEL="embed-multilingual-v3.0",
        EMBEDDING_DIMENSIONS=None,
    )
    gemini_settings = Settings(
        EMBEDDING_PROVIDER="gemini",
        GEMINI_API_KEY="test-key",
        GEMINI_EMBEDDING_MODEL="text-embedding-004",
        EMBEDDING_DIMENSIONS=None,
    )
    ollama_settings = Settings(
        EMBEDDING_PROVIDER="ollama",
        OLLAMA_BASE_URL="http://localhost:11434",
        OLLAMA_EMBEDDING_MODEL="nomic-embed-text",
        EMBEDDING_DIMENSIONS=None,
    )

    async with httpx.AsyncClient() as client:
        cohere_adapter = create_embedding_adapter(cohere_settings, http_client=client)
        gemini_adapter = create_embedding_adapter(gemini_settings, http_client=client)
        ollama_adapter = create_embedding_adapter(ollama_settings, http_client=client)

    assert cohere_adapter.dimensions == DEFAULT_COHERE_DIMENSIONS
    assert gemini_adapter.dimensions == DEFAULT_OTHERS_DIMENSIONS
    assert ollama_adapter.dimensions == DEFAULT_OTHERS_DIMENSIONS
