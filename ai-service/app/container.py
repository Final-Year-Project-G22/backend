from __future__ import annotations

from typing import TYPE_CHECKING

from dependency_injector import containers, providers

from app.config import Settings
from core.ports.cache import CachePort
from core.ports.core_service import CoreServicePort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.llm import LLMPort
from core.usecases import AskAIUseCase, ConversationUseCase, QuotaGuardUseCase
from infrastructure.database.connection import async_session_factory
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyKnowledgeRepository,
    SqlAlchemyQuotaRepository,
)

if TYPE_CHECKING:
    pass


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

    embedding_port: providers.Dependency[EmbeddingPort] = providers.Dependency(
        instance_of=EmbeddingPort,
    )
    llm_port: providers.Dependency[LLMPort] = providers.Dependency(
        instance_of=LLMPort,
    )
    cache_port: providers.Dependency[CachePort] = providers.Dependency(
        instance_of=CachePort,
    )
    event_bus_port: providers.Dependency[EventBusPort] = providers.Dependency(
        instance_of=EventBusPort,
    )
    core_service_port: providers.Dependency[CoreServicePort] = providers.Dependency(
        instance_of=CoreServicePort,
    )

    # Use-case providers
    quota_guard = providers.Factory(
        QuotaGuardUseCase,
        quota_repository=quota_repository,
        core_service=core_service_port,
    )
    conversation = providers.Factory(
        ConversationUseCase,
        conversation_repository=conversation_repository,
        quota_guard=quota_guard,
    )
    ask_ai = providers.Factory(
        AskAIUseCase,
        conversation=conversation,
        quota_guard=quota_guard,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
        cache=cache_port,
        event_bus=event_bus_port,
    )
