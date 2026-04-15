from __future__ import annotations

from unittest.mock import AsyncMock

import pytest
from sqlalchemy.ext.asyncio import AsyncSession

from app.container import Container
from core.ports.cache import CachePort
from core.ports.core_service import CoreServicePort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.llm import LLMPort
from core.usecases import AskAIUseCase, ConversationUseCase, QuotaGuardUseCase
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyKnowledgeRepository,
    SqlAlchemyQuotaRepository,
)
from infrastructure.messagebus import IngestionRequestedConsumer
from workers.tasks import IngestionRequestedTaskHandler


@pytest.mark.asyncio
async def test_container_wires_database_repositories() -> None:
    container = Container()

    quota_repository = container.quota_repository()
    conversation_repository = container.conversation_repository()
    knowledge_repository = container.knowledge_repository()

    assert isinstance(quota_repository, SqlAlchemyQuotaRepository)
    assert isinstance(conversation_repository, SqlAlchemyConversationRepository)
    assert isinstance(knowledge_repository, SqlAlchemyKnowledgeRepository)
    assert isinstance(quota_repository.session, AsyncSession)
    assert isinstance(conversation_repository.session, AsyncSession)
    assert isinstance(knowledge_repository.session, AsyncSession)

    await quota_repository.session.close()
    await conversation_repository.session.close()
    await knowledge_repository.session.close()


def test_container_wires_use_cases_with_dependency_overrides() -> None:
    container = Container(
        embedding_port=AsyncMock(spec=EmbeddingPort),
        llm_port=AsyncMock(spec=LLMPort),
        cache_port=AsyncMock(spec=CachePort),
        event_bus_port=AsyncMock(spec=EventBusPort),
        core_service_port=AsyncMock(spec=CoreServicePort),
    )

    quota_guard = container.quota_guard()
    conversation = container.conversation()
    ask_ai = container.ask_ai()

    assert isinstance(quota_guard, QuotaGuardUseCase)
    assert isinstance(conversation, ConversationUseCase)
    assert isinstance(ask_ai, AskAIUseCase)


def test_container_wires_ingestion_worker_dependencies() -> None:
    container = Container()

    task_handler = container.ingestion_requested_task_handler()
    consumer = container.ingestion_consumer()

    assert isinstance(task_handler, IngestionRequestedTaskHandler)
    assert isinstance(consumer, IngestionRequestedConsumer)
