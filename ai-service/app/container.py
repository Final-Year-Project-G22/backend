from __future__ import annotations

from dependency_injector import containers, providers

from app.config import Settings


class Container(containers.DeclarativeContainer):
    wiring_config = containers.WiringConfiguration(
        modules=["main", "api.v1.router"],
    )

    config = providers.Singleton(Settings)

    # Port providers — concrete adapters wired in Phase 3
    # embedding_port = providers.Factory(CohereEmbeddingAdapter)
    # llm_port = providers.Factory(GeminiProAdapter)
    # repository_port = providers.Factory(PgVectorRepository)
    # cache_port = providers.Factory(RedisCacheAdapter)
    # event_bus_port = providers.Singleton(RabbitMQBus)
