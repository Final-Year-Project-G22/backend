from __future__ import annotations

from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
    )

    # Application
    APP_NAME: str = "ai-service"
    APP_VERSION: str = "0.1.0"
    DEBUG: bool = False
    LOG_LEVEL: str = "INFO"

    # HTTP
    HTTP_PORT: int = 8000
    CORS_ORIGINS: list[str] = ["http://localhost:3000"]

    # gRPC
    GRPC_PORT: int = 50051
    CORE_GRPC_ENDPOINT: str = "localhost:50052"

    # Database
    DATABASE_URL: str = "postgresql+asyncpg://postgres:postgres@localhost:5432/adisu_ai"

    # Cache
    REDIS_URL: str = "redis://localhost:6379/0"

    # Message Bus
    RABBITMQ_URL: str = "amqp://guest:guest@localhost:5672/"

    # Embeddings
    EMBEDDING_PROVIDER: str = "cohere"
    COHERE_API_KEY: str = ""
    GEMINI_API_KEY: str = ""
    COHERE_EMBEDDING_MODEL: str = "embed-multilingual-v3.0"
    GEMINI_EMBEDDING_MODEL: str = "text-embedding-004"
    OLLAMA_EMBEDDING_MODEL: str = "nomic-embed-text"
    OLLAMA_BASE_URL: str = "http://localhost:11434"
    EMBEDDING_DIMENSIONS: int | None = None

    # LLM
    LLM_PROVIDER: str = "gemini"
    COHERE_LLM_MODEL: str = "command-r"
    GEMINI_LLM_MODEL: str = "gemini-1.5-flash"
    OLLAMA_LLM_MODEL: str = "qwen2.5"

    # Observability
    OTEL_ENABLED: bool = False
    PROMETHEUS_PORT: int = 9090


@lru_cache
def get_settings() -> Settings:
    return Settings()
