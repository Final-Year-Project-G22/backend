import httpx

from app.config import Settings
from core.domain.exceptions import ConfigurationError
from core.ports.llm import LLMPort
from infrastructure.llm.cohere import CohereLLMAdapter
from infrastructure.llm.gemini import GeminiLLMAdapter
from infrastructure.llm.ollama import OllamaLLMAdapter


def create_llm_adapter(
    settings: Settings,
    *,
    http_client: httpx.AsyncClient,
) -> LLMPort:
    provider = settings.LLM_PROVIDER.strip().lower()

    if provider == "cohere":
        return CohereLLMAdapter(
            api_key=settings.COHERE_API_KEY,
            model=settings.COHERE_LLM_MODEL,
            http_client=http_client,
        )

    if provider == "gemini":
        return GeminiLLMAdapter(
            api_key=settings.GEMINI_API_KEY,
            model=settings.GEMINI_LLM_MODEL,
            http_client=http_client,
            use_vertex=settings.GEMINI_USE_VERTEX,
            vertex_project=settings.GEMINI_VERTEX_PROJECT,
            vertex_location=settings.GEMINI_VERTEX_LOCATION,
        )

    if provider == "ollama":
        return OllamaLLMAdapter(
            base_url=settings.OLLAMA_BASE_URL,
            model=settings.OLLAMA_LLM_MODEL,
            http_client=http_client,
        )

    raise ConfigurationError(
        "unsupported llm provider",
        details={"provider": settings.LLM_PROVIDER},
    )


__all__ = [
    "CohereLLMAdapter",
    "GeminiLLMAdapter",
    "OllamaLLMAdapter",
    "create_llm_adapter",
]
