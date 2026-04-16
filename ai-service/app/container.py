from __future__ import annotations

from typing import cast

from dependency_injector import containers, providers

from app.config import Settings
from app.security import build_ingestion_envelope_verifier
from core.ports.cache import CachePort
from core.ports.core_service import CoreServicePort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.llm import LLMPort
from core.usecases import AskAIUseCase, ConversationUseCase, QuotaGuardUseCase
from infrastructure.database.connection import async_session_factory
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyIngestionEventLedgerRepository,
    SqlAlchemyKnowledgeRepository,
    SqlAlchemyQuotaRepository,
)
from infrastructure.messagebus import IngestionRequestedConsumer
from infrastructure.rpc import CoreServiceGrpcAdapter, CoreUserGrpcClient
from workers.tasks import IngestionRequestedTaskHandler


class Container(containers.DeclarativeContainer):
    wiring_config = containers.WiringConfiguration(
        modules=["main", "workers.ingestion_worker"],
    )

    config = providers.Singleton(Settings)

    core_grpc_client = providers.Singleton(
        CoreUserGrpcClient,
        endpoint=config.provided.CORE_GRPC_ENDPOINT,
    )

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
    ingestion_event_ledger_repository = providers.Factory(
        SqlAlchemyIngestionEventLedgerRepository,
        session=db_session,
    )

    embedding_port: providers.Dependency[EmbeddingPort] = providers.Dependency(
        instance_of=EmbeddingPort,
    )
    llm_port: providers.Dependency[LLMPort] = providers.Dependency(
        instance_of=LLMPort,
    )

    cache_port: providers.Provider[CachePort | None] = cast(
        providers.Provider[CachePort | None],
        providers.Object(None),
    )
    event_bus_port: providers.Provider[EventBusPort | None] = cast(
        providers.Provider[EventBusPort | None],
        providers.Object(None),
    )
    core_service_port: providers.Provider[CoreServicePort | None] = cast(
        providers.Provider[CoreServicePort | None],
        providers.Singleton(
            CoreServiceGrpcAdapter,
            endpoint=config.provided.CORE_GRPC_ENDPOINT,
            client=core_grpc_client,
        ),
    )

    ingestion_envelope_verifier = providers.Factory(
        build_ingestion_envelope_verifier,
        settings=config,
    )
    ingestion_requested_task_handler = providers.Factory(
        IngestionRequestedTaskHandler,
        envelope_verifier=ingestion_envelope_verifier,
        ingestion_event_ledger_repository=ingestion_event_ledger_repository,
    )
    ingestion_consumer = providers.Factory(
        IngestionRequestedConsumer,
        rabbitmq_url=config.provided.RABBITMQ_URL,
        queue_name=config.provided.INGESTION_WORKER_QUEUE,
        exchange_name=config.provided.INGESTION_WORKER_EXCHANGE,
        routing_key=config.provided.INGESTION_WORKER_ROUTING_KEY,
        prefetch_count=config.provided.INGESTION_WORKER_PREFETCH_COUNT,
        requeue_on_failure=config.provided.INGESTION_WORKER_REQUEUE_ON_FAILURE,
    )

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
