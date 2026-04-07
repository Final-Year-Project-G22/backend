from __future__ import annotations

from dependency_injector import containers, providers

from app.config import Settings
from infrastructure.database.connection import async_session_factory
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyKnowledgeRepository,
    SqlAlchemyQuotaRepository,
)


class Container(containers.DeclarativeContainer):
    wiring_config = containers.WiringConfiguration(
        modules=["main"],
    )

    config = providers.Singleton(Settings)

    db_session = providers.Callable(async_session_factory)

    quota_repository = providers.Factory(
        SqlAlchemyQuotaRepository,
        session=db_session,
    )
    conversation_repository = providers.Factory(
        SqlAlchemyConversationRepository,
        session=db_session,
    )
    knowledge_repository = providers.Factory(
        SqlAlchemyKnowledgeRepository,
        session=db_session,
    )

    # Port providers — concrete adapters wired in Phase 3
    # embedding_port = providers.Factory(CohereEmbeddingAdapter)
    # llm_port = providers.Factory(GeminiProAdapter)
    # repository_port = providers.Factory(PgVectorRepository)
    # cache_port = providers.Factory(RedisCacheAdapter)
    # event_bus_port = providers.Singleton(RabbitMQBus)
