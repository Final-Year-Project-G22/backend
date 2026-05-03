from __future__ import annotations

from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
    )

    APP_NAME: str = "ai-service"
    APP_VERSION: str = "0.1.0"
    DEBUG: bool = False
    LOG_LEVEL: str = "INFO"

    HTTP_PORT: int = 8000
    CORS_ORIGINS: list[str] = ["http://localhost:3000"]
    # When True, httpx honors HTTP(S)_PROXY from the environment (default httpx behavior).
    HTTPX_TRUST_ENV: bool = True

    GRPC_PORT: int = 50051
    CORE_GRPC_ENDPOINT: str = "localhost:50052"
    INTERNAL_GRPC_AUTH_TOKEN: str = ""

    SEAWEEDFS_FILER_URL: str = "http://localhost:8888"

    DATABASE_URL: str = "postgresql+asyncpg://postgres:postgres@localhost:5432/adisu_ai"

    REDIS_URL: str = "redis://localhost:6379/0"

    RABBITMQ_URL: str = "amqp://guest:guest@localhost:5672/"

    INGESTION_WORKER_QUEUE: str = "ai.ingestion.requested.v1"
    INGESTION_WORKER_EXCHANGE: str = "core.events"
    INGESTION_WORKER_ROUTING_KEY: str = "document.ingestion.requested.v1"
    INGESTION_WORKER_PREFETCH_COUNT: int = 16
    INGESTION_WORKER_REQUEUE_ON_FAILURE: bool = True
    INGESTION_WORKER_MAX_RETRIES: int = 3
    INGESTION_WORKER_RETRY_BASE_DELAY_MS: int = 1000
    INGESTION_WORKER_RETRY_MAX_DELAY_MS: int = 30000
    INGESTION_WORKER_RETRY_JITTER_FACTOR: float = 0.1

    INGESTION_WORKER_DLQ_ENABLED: bool = True
    INGESTION_WORKER_DLQ_EXCHANGE: str = "ai.ingestion.dlq"
    INGESTION_WORKER_DLQ_ROUTING_KEY: str = "document.ingestion.dlq.v1"

    # Ingestion envelope verification
    INGESTION_SIGNING_ACTIVE_KEY_ID: str = "ingestion-v1"
    INGESTION_SIGNING_ACTIVE_KEY_SECRET: str = "change-me"
    INGESTION_SIGNING_PREVIOUS_KEYS_JSON: str = "{}"

    # Embeddings
    EMBEDDING_PROVIDER: str = "cohere"
    COHERE_API_KEY: str = ""
    GEMINI_API_KEY: str = ""
    COHERE_EMBEDDING_MODEL: str = "embed-multilingual-v3.0"
    GEMINI_EMBEDDING_MODEL: str = "text-embedding-004"
    OLLAMA_EMBEDDING_MODEL: str = "nomic-embed-text"
    OLLAMA_BASE_URL: str = "http://localhost:11434"
    EMBEDDING_DIMENSIONS: int | None = None

    LLM_PROVIDER: str = "cohere"
    COHERE_LLM_MODEL: str = "command-a-03-2025"
    GEMINI_LLM_MODEL: str = "gemini-1.5-flash"
    OLLAMA_LLM_MODEL: str = "qwen2.5"
    AI_ASK_ENABLED: bool = True

    OTEL_ENABLED: bool = False
    PROMETHEUS_PORT: int = 9090


@lru_cache
def get_settings() -> Settings:
    return Settings()
