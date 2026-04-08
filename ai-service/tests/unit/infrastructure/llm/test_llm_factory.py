from __future__ import annotations

import httpx
import pytest

from app.config import Settings
from core.domain.exceptions import ConfigurationError
from infrastructure.llm import (
    CohereLLMAdapter,
    GeminiLLMAdapter,
    OllamaLLMAdapter,
    create_llm_adapter,
)


@pytest.mark.asyncio
async def test_llm_factory_returns_cohere_adapter() -> None:
    settings = Settings(
        LLM_PROVIDER="cohere",
        COHERE_API_KEY="test-key",
        COHERE_LLM_MODEL="command-r",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_llm_adapter(settings, http_client=client)

    assert isinstance(adapter, CohereLLMAdapter)


@pytest.mark.asyncio
async def test_llm_factory_returns_gemini_adapter() -> None:
    settings = Settings(
        LLM_PROVIDER="gemini",
        GEMINI_API_KEY="test-key",
        GEMINI_LLM_MODEL="gemini-1.5-flash",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_llm_adapter(settings, http_client=client)

    assert isinstance(adapter, GeminiLLMAdapter)


@pytest.mark.asyncio
async def test_llm_factory_returns_ollama_adapter() -> None:
    settings = Settings(
        LLM_PROVIDER="ollama",
        OLLAMA_BASE_URL="http://localhost:11434",
        OLLAMA_LLM_MODEL="qwen2.5",
    )
    async with httpx.AsyncClient() as client:
        adapter = create_llm_adapter(settings, http_client=client)

    assert isinstance(adapter, OllamaLLMAdapter)


@pytest.mark.asyncio
async def test_llm_factory_raises_for_unknown_provider() -> None:
    settings = Settings(LLM_PROVIDER="unknown")
    async with httpx.AsyncClient() as client:
        with pytest.raises(ConfigurationError, match="unsupported llm provider"):
            create_llm_adapter(settings, http_client=client)
