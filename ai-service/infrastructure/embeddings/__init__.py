import httpx

from app.config import Settings
from core.domain.exceptions import ConfigurationError
from core.ports.embedding import EmbeddingPort
from infrastructure.embeddings.cohere import CohereEmbeddingAdapter
from infrastructure.embeddings.gemini import GeminiEmbeddingAdapter
from infrastructure.embeddings.ollama import OllamaEmbeddingAdapter


def create_embedding_adapter(
    settings: Settings,
    *,
    http_client: httpx.AsyncClient,
) -> EmbeddingPort:
    provider = settings.EMBEDDING_PROVIDER.strip().lower()
    dimensions = settings.EMBEDDING_DIMENSIONS

    if provider == "cohere":
        return CohereEmbeddingAdapter(
            api_key=settings.COHERE_API_KEY,
            model=settings.COHERE_EMBEDDING_MODEL,
            dimensions=dimensions or 1024,
            http_client=http_client,
        )

    if provider == "gemini":
        return GeminiEmbeddingAdapter(
            api_key=settings.GEMINI_API_KEY,
            model=settings.GEMINI_EMBEDDING_MODEL,
            dimensions=dimensions or 768,
            http_client=http_client,
            use_vertex=settings.GEMINI_USE_VERTEX,
            vertex_project=settings.GEMINI_VERTEX_PROJECT,
            vertex_location=settings.GEMINI_VERTEX_LOCATION,
        )

    if provider == "ollama":
        return OllamaEmbeddingAdapter(
            base_url=settings.OLLAMA_BASE_URL,
            model=settings.OLLAMA_EMBEDDING_MODEL,
            dimensions=dimensions or 768,
            http_client=http_client,
        )

    raise ConfigurationError(
        "unsupported embedding provider",
        details={"provider": settings.EMBEDDING_PROVIDER},
    )


__all__ = [
    "CohereEmbeddingAdapter",
    "GeminiEmbeddingAdapter",
    "OllamaEmbeddingAdapter",
    "create_embedding_adapter",
]
