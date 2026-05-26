from __future__ import annotations

from typing import Any, cast

from dependency_injector import containers, providers

import grpc as grpc_lib
from app.config import Settings
from app.security import build_ingestion_envelope_verifier
from core.ports.cache import CachePort
from core.ports.core_service import CoreServicePort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.llm import LLMPort
from core.usecases import (
    AskAIUseCase,
    ConversationUseCase,
    IngestionOrchestratorUseCase,
    QuotaGuardUseCase,
)
from core.usecases.strategies.agentic_ask import AgenticAskStrategy
from core.usecases.strategies.simple_ask import SimpleAskStrategy
from infrastructure.chunking.registry import ChunkingRegistry
from infrastructure.database.connection import async_session_factory
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyIngestionEventLedgerRepository,
    SqlAlchemyKnowledgeRepository,
    SqlAlchemyQuotaRepository,
)
from infrastructure.messagebus import IngestionRequestedConsumer
from infrastructure.messagebus.rabbitmq_event_bus import RabbitMQEventBusAdapter
from infrastructure.parsers.registry import ParserRegistry
from infrastructure.prefetch.pipeline import PreFetchPipeline
from infrastructure.prompts import PromptLoader
from infrastructure.rpc import AIToolGrpcClient, CoreServiceGrpcAdapter, CoreUserGrpcClient
from infrastructure.tools.intent_classifier import EmbeddingIntentClassifier
from infrastructure.tools.local.search_knowledge_base import SearchKnowledgeBaseTool
from infrastructure.tools.local.search_trusted_web import SearchTrustedWebTool
from infrastructure.tools.tool_registry import ToolRegistry
from workers.tasks import IngestionRequestedTaskHandler


def _create_document_fetch_channel(endpoint: str) -> Any:
    """Factory for the document-fetch gRPC channel.

    Returns ``Any`` because Pylance cannot resolve ``grpc.aio.Channel``
    from the grpc stubs, which would otherwise make the provider type
    ``Singleton[Unknown]`` and cascade "partially unknown" errors.
    """
    return cast(Any, grpc_lib.aio.insecure_channel(endpoint))  # type: ignore[reportUnknownMemberType]


class Container(containers.DeclarativeContainer):
    wiring_config = containers.WiringConfiguration(
        modules=["main", "workers.ingestion_worker"],
    )

    config = providers.Singleton(Settings)

    prompt_loader = providers.Singleton(
        PromptLoader,
        template_dir=config.provided.AI_PROMPT_DIR,
    )

    core_grpc_client = providers.Singleton(
        CoreUserGrpcClient,
        endpoint=config.provided.CORE_GRPC_ENDPOINT,
    )

    ai_tool_client = providers.Singleton(
        AIToolGrpcClient,
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

    search_knowledge_base_tool = providers.Factory(
        SearchKnowledgeBaseTool,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
    )

    search_trusted_web_tool = providers.Factory(
        SearchTrustedWebTool,
    )

    tool_registry = providers.Singleton(
        ToolRegistry,
        ai_tool_client=ai_tool_client,
        expected_remote_tools=config.provided.AI_EXPECTED_REMOTE_TOOLS,
        search_knowledge_base=search_knowledge_base_tool,
        search_trusted_web=search_trusted_web_tool,
    )

    intent_classifier = providers.Singleton(
        EmbeddingIntentClassifier,
        embedding_port=embedding_port,
        knowledge_seeds=config.provided.AI_INTENT_SEED_QUERIES_KNOWLEDGE,
        personal_seeds=config.provided.AI_INTENT_SEED_QUERIES_PERSONAL,
        threshold=config.provided.AI_INTENT_SIMILARITY_THRESHOLD,
    )

    pre_fetch_pipeline = providers.Singleton(
        PreFetchPipeline,
        tool_registry=tool_registry,
    )

    cache_port: providers.Provider[CachePort | None] = cast(
        providers.Provider[CachePort | None],
        providers.Object(None),
    )
    event_bus_port: providers.Provider[EventBusPort | None] = cast(
        providers.Provider[EventBusPort | None],
        providers.Singleton(
            RabbitMQEventBusAdapter,
            amqp_url=config.provided.RABBITMQ_URL,
            exchange_name=config.provided.INGESTION_WORKER_EXCHANGE,
        ),
    )
    document_fetch_channel: providers.Provider[Any] = cast(
        providers.Provider[Any],
        providers.Singleton(
            _create_document_fetch_channel,
            config.provided.CORE_GRPC_ENDPOINT,
        ),
    )
    core_service_port: providers.Provider[CoreServicePort | None] = cast(
        providers.Provider[CoreServicePort | None],
        providers.Singleton(
            CoreServiceGrpcAdapter,
            endpoint=config.provided.CORE_GRPC_ENDPOINT,
            client=core_grpc_client,
            document_fetch_channel=document_fetch_channel,
        ),
    )

    ingestion_envelope_verifier = providers.Factory(
        build_ingestion_envelope_verifier,
        settings=config,
    )
    parser_registry = providers.Singleton(ParserRegistry)
    chunking_registry = providers.Singleton(ChunkingRegistry)

    ingestion_orchestrator = providers.Factory(
        IngestionOrchestratorUseCase,
        event_bus=event_bus_port,
        parser_registry=parser_registry,
        chunking_registry=chunking_registry,
        embedding_port=embedding_port,
        knowledge_repository=knowledge_repository,
        core_service_port=core_service_port,
        seaweedfs_filer_url=config.provided.SEAWEEDFS_FILER_URL,
    )
    ingestion_requested_task_handler = providers.Factory(
        IngestionRequestedTaskHandler,
        envelope_verifier=ingestion_envelope_verifier,
        ingestion_event_ledger_repository=ingestion_event_ledger_repository,
        ingestion_orchestrator=ingestion_orchestrator,
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

    simple_ask_strategy = providers.Factory(
        SimpleAskStrategy,
        conversation=conversation,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
        prompt_loader=prompt_loader,
        tool_registry=tool_registry,
        cache=cache_port,
        event_bus=event_bus_port,
    )

    agentic_ask_strategy = providers.Factory(
        AgenticAskStrategy,
        conversation=conversation,
        knowledge_repository=knowledge_repository,
        embedding_port=embedding_port,
        llm_port=llm_port,
        prompt_loader=prompt_loader,
        tool_registry=tool_registry,
        intent_classifier=intent_classifier,
        pre_fetch_pipeline=pre_fetch_pipeline,
        max_iterations=config.provided.AI_AGENTIC_MAX_ITERATIONS,
        cache=cache_port,
        event_bus=event_bus_port,
    )

    ask_ai = providers.Factory(
        AskAIUseCase,
        simple_strategy=simple_ask_strategy,
        agentic_strategy=agentic_ask_strategy,
        agentic_enabled=config.provided.AI_AGENTIC_ENABLED,
    )
